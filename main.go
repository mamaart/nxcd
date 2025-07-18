package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mamaart/nxcd/notifier"
)

func main() {
	cfg := Config{
		PrivateKeyPath: os.Getenv("SSH_PRIVATE_KEY_PATH"),
		Repo:           os.Getenv("REPO"),
		Branch:         os.Getenv("BRANCH"),
		Host:           os.Getenv("HOST"),
	}

	pollDuration := os.Getenv("POLL_DURATION")
	if pollDuration != "" {
		d, err := strconv.Atoi(pollDuration)
		if err == nil {
			cfg.PollDuration = time.Second * time.Duration(d)
		}
	}

	if enabled := os.Getenv("MATRIX_ENABLED"); enabled == "true" {
		n, err := notifier.FromEnv()
		if err != nil {
			log.Fatal(err)
		}
		cfg.Notifier = n.Notify
	} else {
		cfg.Notifier = func(s string) {}
	}

	log.Fatal(Run(cfg))
}
