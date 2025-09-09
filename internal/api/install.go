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

	if err := ApplyBaseConfigDefaults(); err != nil {
		return err
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

// ApplyBaseConfigDefaults reads the base-config.json, applies default values,
// validates the configuration, and writes it back to the file.
func ApplyBaseConfigDefaults() error {
	conf, err := config.ReadConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}

	// Set default values if they are not already set
	if conf.System.Username == "" {
		username, err := config.GetUsername()
		if err != nil {
			return fmt.Errorf("error getting username: %w", err)
		}
		conf.System.Username = username
	}
	if conf.System.Desktop == "" {
		desktop, err := config.GetDesktop()
		if err != nil {
			return fmt.Errorf("error getting desktop: %w", err)
		}
		conf.System.Desktop = desktop
	}
	if conf.System.Type == "" {
		systemType, err := config.GetType()
		if err != nil {
			return fmt.Errorf("error getting system type: %w", err)
		}
		conf.System.Type = systemType
	}

	// Validate that at least one username matches system.username
	userMatch := false
	for _, user := range conf.Users {
		if user.Username == conf.System.Username {
			userMatch = true
			break
		}
	}

	// Validate users array
	if !userMatch || len(conf.Users) == 0 {
		// return fmt.Errorf("the 'users' array in base-config.json cannot be empty")
		conf.Users = append(conf.Users, config.User{
			Username: conf.System.Username,
			Email:    fmt.Sprintf("%s@pilo", conf.System.Username),
			Name:     conf.System.Username,
		})
	}

	// if !userMatch {
	// 	return fmt.Errorf("at least one user in base-config.json must match the system username ('%s')", conf.System.Username)
	// }

	// Write the updated config back to the file
	if err := config.WriteConfig(conf); err != nil {
		return fmt.Errorf("error writing updated config: %w", err)
	}

	return nil
}
