package config

import (
	"errors"
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

func (c Config) Valid() error {
	if c.Matrix.Enabled {
		if c.Matrix.HomeServer == "" {
			return errors.New("missing matrix homeserver address")
		}
		if c.Matrix.Username == "" {
			return errors.New("missing matrix user id")
		}
		if c.Matrix.Password == "" {
			return errors.New("missing matrix token")
		}
		if c.Matrix.RoomID == "" {
			return errors.New("missing matrix room id")
		}
	}
	if c.Git.Branch == "" {
		c.Git.Branch = "main"
	}
	if c.Git.PrivateKeyPath == "" {
		return errors.New("missing private_key_path")
	}
	if c.Git.Repo == "" {
		return errors.New("missing github repo in the format username/repo")
	}
	if c.NixHost == "" {
		return errors.New("missing nix machine hostname")
	}
	if c.PollDuration == "" {
		c.PollDuration = "60"
	}
	return nil
}

func FromFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
