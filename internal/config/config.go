package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const (
	envAPIID   = "TMGC_API_ID"
	envAPIHash = "TMGC_API_HASH"
	envStore   = "TMGC_SESSION_STORE"
)

type Config struct {
	APIID        int    `json:"api_id"`
	APIHash      string `json:"api_hash"`
	SessionStore string `json:"session_store"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}

func (c *Config) ApplyEnv() error {
	if v := os.Getenv(envAPIID); v != "" {
		id, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("%s must be int: %w", envAPIID, err)
		}
		c.APIID = id
	}
	if v := os.Getenv(envAPIHash); v != "" {
		c.APIHash = v
	}
	if v := os.Getenv(envStore); v != "" {
		c.SessionStore = v
	}
	return nil
}
