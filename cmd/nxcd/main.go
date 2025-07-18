package main

import (
	"log"
	"time"

	"github.com/mamaart/nxcd/internal/config"
	"github.com/mamaart/nxcd/internal/core"
	"github.com/mamaart/nxcd/notifier"
)

func main() {
	config, err := config.Load()
	if err != nil {
		log.Fatalf("invalid config: %v", err)
	}
	cfg := core.Config{
		PrivateKeyPath: config.Git.PrivateKeyPath,
		Repo:           config.Git.Repo,
		Branch:         config.Git.Branch,
		Host:           config.NixHost,
		PollDuration:   time.Second * time.Duration(config.PollDuration),
		Notifier:       func(s string) {},
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
	}
	log.Fatal(core.Run(cfg))
}
