package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mamaart/nxcd/internal/config"
	"github.com/mamaart/nxcd/notifier"
)

func main() {
	config := loadConfig()
	if err := config.Valid(); err != nil {
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

func loadConfig() *config.Config {
	configPath := os.Getenv("APP_CONFIG")
	if configPath != "" {
		config, err := config.FromFile(configPath)
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
		return config
	}
	config := config.Config{
		Matrix: config.Matrix{
			Enabled:    os.Getenv("MATRIX_ENABLED") == "true",
			HomeServer: os.Getenv("MATRIX_HOMESERVER"),
			Username:   os.Getenv("MATRIX_USERNAME"),
			Password:   os.Getenv("MATRIX_PASSWORD"),
			RoomID:     os.Getenv("MATRIX_ROOMID"),
		},
		NixHost: os.Getenv("NIX_HOST"),
		Git: config.Git{
			Repo:           os.Getenv("GIT_REPO"),
			PrivateKeyPath: os.Getenv("GIT_SSH_PRIVATE_KEY_PATH"),
			Branch:         os.Getenv("GIT_BRANCH"),
		},
	}
	return &config
}
