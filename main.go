package main

import (
	"log"
	"strconv"
	"time"

	"github.com/mamaart/nxcd/internal/config"
	"github.com/mamaart/nxcd/notifier"
)

func main() {
	config, err := config.Load()
	if err != nil {
		log.Fatalf("invalid config: %v", err)
	}
	cfg := Config{
		PrivateKeyPath: config.Git.PrivateKeyPath,
		Repo:           config.Git.Repo,
		Branch:         config.Git.Branch,
		Host:           config.NixHost,
	}

	d, err := strconv.Atoi(config.PollDuration)
	if err == nil {
		cfg.PollDuration = time.Second * time.Duration(d)
	}

	if config.Matrix.Enabled {
		n, err := notifier.New(
			config.Matrix.HomeServer,
			config.Matrix.Username,
			config.Matrix.Password,
			config.Matrix.RoomID,
		)
		if err != nil {
			log.Fatal(err)
		}
		cfg.Notifier = n.Notify
	} else {
		cfg.Notifier = func(s string) {}
	}

	log.Fatal(Run(cfg))
}
