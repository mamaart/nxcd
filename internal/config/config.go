package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	NixHost      string `yaml:"nix_host"`
	Git          Git    `yaml:"git"`
	Matrix       Matrix `yaml:"matrix"`
	PollDuration string `yaml:"poll_duration"`
}

type Matrix struct {
	Enabled    bool   `yaml:"enabled"`
	HomeServer string `yaml:"home_server"`
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	RoomID     string `yaml:"room_id"`
}

type Git struct {
	Repo           string `yaml:"repo"`
	Branch         string `yaml:"branch"`
	PrivateKeyPath string `yaml:"private_key_path"`
}

func Load() (Config, error) {
	configPath := os.Getenv("APP_CONFIG")
	if configPath == "" {
		return Config{}, fmt.Errorf("missing APP_CONFIG environment variable")
	}
	if configPath != "" {
		config, err := fromFile(configPath)
		if err != nil {
			return Config{}, fmt.Errorf("Failed to load configuration: %v", err)
		}
		return config.valid()
	}
	config := Config{
		Matrix: Matrix{
			Enabled:    os.Getenv("MATRIX_ENABLED") == "true",
			HomeServer: os.Getenv("MATRIX_HOMESERVER"),
			Username:   os.Getenv("MATRIX_USERNAME"),
			Password:   os.Getenv("MATRIX_PASSWORD"),
			RoomID:     os.Getenv("MATRIX_ROOMID"),
		},
		NixHost: os.Getenv("NIX_HOST"),
		Git: Git{
			Repo:           os.Getenv("GIT_REPO"),
			PrivateKeyPath: os.Getenv("GIT_SSH_PRIVATE_KEY_PATH"),
			Branch:         os.Getenv("GIT_BRANCH"),
		},
	}
	return config.valid()
}

func (c Config) valid() (Config, error) {
	if c.Matrix.Enabled {
		if c.Matrix.HomeServer == "" {
			return c, errors.New("missing matrix homeserver address")
		}
		if c.Matrix.Username == "" {
			return c, errors.New("missing matrix user id")
		}
		if c.Matrix.Password == "" {
			return c, errors.New("missing matrix token")
		}
		if c.Matrix.RoomID == "" {
			return c, errors.New("missing matrix room id")
		}
	}
	if c.Git.Repo == "" {
		return c, errors.New("missing github repo in the format username/repo")
	}
	if c.NixHost == "" {
		return c, errors.New("missing nix machine hostname")
	}
	if c.PollDuration == "" {
		c.PollDuration = "60"
	}
	if c.Git.Branch == "" {
		c.Git.Branch = "main"
	}
	if c.Git.PrivateKeyPath == "" {
		c.Git.PrivateKeyPath = "/etc/ssh/ssh_host_ed25519_key"
	}
	return c, nil
}

func fromFile(path string) (Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
