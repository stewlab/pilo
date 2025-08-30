package api

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"pilo/internal/config"
	"pilo/internal/nix"
)

// EnsureNixInstalled checks if Nix is installed and, if not, prompts the user to install it.
func EnsureNixInstalled() error {
	if nix.GetNixMode() == nix.None {
		installCmd := config.GetNixInstallCmd()
		cmd := exec.Command("sh", "-c", installCmd)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Nix installation failed: %w", err)
		}
		fmt.Println("Nix installed successfully. Please restart your shell for the changes to take effect.")
	}
	return nil
}

// CopyNixOSConfigs copies the NixOS configuration files from /etc/nixos to the flake.
func CopyNixOSConfigs(path string) error {
	configs := []string{"hardware-configuration.nix"}
	for _, config := range configs {
		source := filepath.Join("/etc/nixos", config)
		dest := filepath.Join(path, "flake", "hosts", "nixos", config)

		if _, err := os.Stat(source); err == nil {
			if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
				return fmt.Errorf("failed to create destination directory for %s: %w", config, err)
			}
			data, err := os.ReadFile(source)
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", config, err)
			}
			if err := os.WriteFile(dest, data, 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", config, err)
			}
		}
	}
	return nil
}
