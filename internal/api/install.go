package api

import (
	"fmt"
	"os"
	"path/filepath"
	"pilo/internal/config"
	"pilo/internal/nix"

	"fyne.io/fyne/v2"
)

// AutoInstall performs the Pilo configuration installation.
func AutoInstall(path string, win fyne.Window) error {
	registry := config.GetRegistryName()
	remoteURL, _ := config.GetRemoteUrl()

	if err := InstallPilo(path, registry, remoteURL); err != nil {
		return err
	}

	return nil
}

// AutoInstallCLI performs the Pilo configuration installation for the CLI.
func AutoInstallCLI(path string) error {
	registry := config.GetRegistryName()
	remoteURL, _ := config.GetRemoteUrl()

	if err := InstallPilo(path, registry, remoteURL); err != nil {
		return err
	}

	if err := GitAdd(path); err != nil {
		return err
	}
	return GitCommit(path, "pilo: initial commit")
}

// Install performs the Pilo configuration installation.
func InstallPilo(path string, registry string, remoteURL string) error {
	// Ensure the installation path and a default base config exist before inflating
	flakePath := filepath.Join(path, "flake")
	if err := os.MkdirAll(flakePath, 0755); err != nil {
		return fmt.Errorf("error creating flake directory: %w", err)
	}

	configPath := filepath.Join(flakePath, "base-config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := []byte("{\n  \"commit_triggers\": [],\n  \"remote_url\": \"\",\n  \"push_on_commit\": false\n}\n")
		if err := os.WriteFile(configPath, defaultConfig, 0644); err != nil {
			return fmt.Errorf("error creating default base-config.json: %w", err)
		}
	}

	if err := GitInit(path); err != nil {
		return fmt.Errorf("error initializing git repository: %w", err)
	}
	if err := Inflate(path, remoteURL, false); err != nil {
		return fmt.Errorf("error inflating pilo flake: %w", err)
	}

	// If on NixOS, copy the system's configuration files.
	if nix.GetNixMode() == nix.NixOS && remoteURL == "" {
		if err := CopyNixOSConfigs(path); err != nil {
			return fmt.Errorf("error copying NixOS configuration: %w", err)
		}
	}

	// Only attempt to register the flake with Nix if the 'nix' command is available.
	if nix.IsNixInstalled() {
		if err := InstallConfig(path, registry, ""); err != nil {
			return fmt.Errorf("error installing configuration: %w", err)
		}
	}

	return nil
}
