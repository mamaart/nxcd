package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	Idle    = 0
	Running = 1
	Pending = 2
)

type Config struct {
	PrivateKeyPath string
	Repo           string
	Branch         string
	Host           string
	PollDuration   time.Duration
	Notifier       func(string)
}

func Run(cfg Config) error {
	if cfg.Host == "" {
		return fmt.Errorf("missing nix host name for flake")
	}

	repoPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+\/[a-zA-Z0-9._-]+$`)
	if !repoPattern.MatchString(cfg.Repo) {
		return fmt.Errorf("invalid repo, need 'username/reponame'")
	}

	if cfg.PrivateKeyPath == "" {
		cfg.PrivateKeyPath = "/etc/ssh/ssh_host_ed25519_key"
	}
	if cfg.Branch == "" {
		cfg.Branch = "main"
	}
	if cfg.PollDuration == time.Duration(0) {
		cfg.PollDuration = time.Minute
	}

	key, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read file of privateKey: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return fmt.Errorf("failed to parse ssh privateKey: %v", err)
	}

	client, err := ssh.Dial("tcp", "github.com:22", &ssh.ClientConfig{
		User:            "git",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return fmt.Errorf("failed to dial github: %v", err)
	}

	var state atomic.Uint32
	for hash := range loop(client, cfg.Repo, cfg.PollDuration) {
		cfg.Notifier(fmt.Sprintf("new commit on repo: %s", hash))
		switch state.Load() {
		case Idle:
			fmt.Println("Idle")
			if state.CompareAndSwap(Idle, Running) {
				go func(hash string) {
					for {
						runCommand(cfg.Repo, hash, cfg.Host, cfg.Branch)
						if !state.CompareAndSwap(Pending, Running) {
							state.Store(Idle)
							return
						}
					}
				}(hash)
			}
		case Running:
			fmt.Println("Running")
			state.Store(Pending)
		case Pending:
			fmt.Println("Pending")
			// Already pending
		}
	}
	return errors.New("end of loop")
}

func runCommand(repo, hash, host, branch string) {
	cmd := exec.Command(
		"nixos-rebuild", "switch", "--flake",
		fmt.Sprintf(
			"git+ssh://git@github.com/%s?ref=%s&rev=%s#%s",
			repo, branch, hash, host,
		),
	)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if err := cmd.Run(); err != nil {
		log.Printf("failed to run nixos rebuild switch command: %v", err)
	}
}

func loop(client *ssh.Client, repo string, d time.Duration) <-chan string {
	ch := make(chan string)
	go func(ch chan<- string) {
		defer close(ch)

		old, err := getSha(client, repo)
		if err != nil {
			log.Printf("failed to get sha: %v\n", err)
			return
		}

		for range time.Tick(d) {
			sha, err := getSha(client, repo)
			if err != nil {
				log.Printf("failed to get sha: %v\n", err)
				return
			}

			if old != sha {
				ch <- sha
				old = sha
			}
		}
	}(ch)
	return ch
}

func getSha(client *ssh.Client, repo string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("could not get session: %v", err)
	}
	defer session.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("could not get filedescriptor: %v", err)
	}

	if err := session.Start(fmt.Sprintf("git-upload-pack '%s'", repo)); err != nil {
		return "", fmt.Errorf("failed to start session: %v", err)

	}

	hash, err := getSecondLineHash(bufio.NewReader(stdout))
	if err != nil {
		return "", fmt.Errorf("failed to read hash from git response: %v", err)
	}
	return hash, nil
}

func getSecondLineHash(r io.Reader) (string, error) {
	br := bufio.NewReader(r)
	for i := range 2 {
		hdr := make([]byte, 4)
		if _, err := io.ReadFull(br, hdr); err != nil {
			return "", err
		}
		l := 0
		fmt.Sscanf(string(hdr), "%x", &l)
		if l < 4 {
			return "", fmt.Errorf("invalid length")
		}
		data := make([]byte, l-4)
		if _, err := io.ReadFull(br, data); err != nil {
			return "", err
		}
		if i == 1 {
			line := strings.SplitN(string(data), "\x00", 2)[0]
			return strings.Fields(line)[0], nil
		}
	}
	return "", fmt.Errorf("no second line")
}
