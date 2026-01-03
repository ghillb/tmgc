package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Paths struct {
	Root        string
	Profile     string
	ProfileDir  string
	ConfigPath  string
	SessionPath string
	PeersPath   string
}

func ResolvePaths(configPath, profile string) (Paths, error) {
	if profile == "" {
		profile = "default"
	}

	if configPath != "" {
		profileDir := filepath.Dir(configPath)
		return Paths{
			Root:        profileDir,
			Profile:     profile,
			ProfileDir:  profileDir,
			ConfigPath:  configPath,
			SessionPath: filepath.Join(profileDir, "session.json"),
			PeersPath:   filepath.Join(profileDir, "peers.json"),
		}, nil
	}

	root, err := defaultRoot()
	if err != nil {
		return Paths{}, err
	}

	profileDir := filepath.Join(root, "profiles", profile)
	return Paths{
		Root:        root,
		Profile:     profile,
		ProfileDir:  profileDir,
		ConfigPath:  filepath.Join(profileDir, "config.json"),
		SessionPath: filepath.Join(profileDir, "session.json"),
		PeersPath:   filepath.Join(profileDir, "peers.json"),
	}, nil
}

func EnsureDirs(paths Paths) error {
	if paths.ProfileDir == "" {
		return errors.New("invalid profile directory")
	}
	return os.MkdirAll(paths.ProfileDir, 0o700)
}

func defaultRoot() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}
	return filepath.Join(base, "tmgc"), nil
}
