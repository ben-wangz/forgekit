package secret

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	PrivateKeyPath string
	PublicKeyPath  string
}

func loadConfig() (*Config, error) {
	cfg := &Config{}

	if privKey := os.Getenv("SECRET_PRIVATE_KEY"); privKey != "" {
		cfg.PrivateKeyPath = privKey
	}
	if pubKey := os.Getenv("SECRET_PUBLIC_KEY"); pubKey != "" {
		cfg.PublicKeyPath = pubKey
	}

	if cfg.PrivateKeyPath == "" || cfg.PublicKeyPath == "" {
		detected, err := detectSSHKeys()
		if err != nil {
			return nil, err
		}
		if cfg.PrivateKeyPath == "" {
			cfg.PrivateKeyPath = detected.PrivateKeyPath
		}
		if cfg.PublicKeyPath == "" {
			cfg.PublicKeyPath = detected.PublicKeyPath
		}
	}

	if _, err := os.Stat(cfg.PrivateKeyPath); err != nil {
		return nil, fmt.Errorf("private key not found: %s", cfg.PrivateKeyPath)
	}
	if _, err := os.Stat(cfg.PublicKeyPath); err != nil {
		return nil, fmt.Errorf("public key not found: %s", cfg.PublicKeyPath)
	}

	return cfg, nil
}

func detectSSHKeys() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot get home directory: %w", err)
	}

	sshDir := filepath.Join(home, ".ssh")
	privKey := filepath.Join(sshDir, "id_ed25519")
	pubKey := filepath.Join(sshDir, "id_ed25519.pub")

	if _, err := os.Stat(privKey); err != nil {
		return nil, fmt.Errorf("ed25519 private key not found at %s", privKey)
	}
	if _, err := os.Stat(pubKey); err != nil {
		return nil, fmt.Errorf("ed25519 public key not found at %s", pubKey)
	}

	return &Config{PrivateKeyPath: privKey, PublicKeyPath: pubKey}, nil
}
