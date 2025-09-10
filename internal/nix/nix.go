package nix

import (
	"fmt"
	"os"
	"os/exec"
	"pilo/internal/config"
	"strings"
)

type NixMode string

const (
	NixOS      NixMode = "NixOS"
	NonNixOS   NixMode = "Non-NixOS"
	None       NixMode = "None"
	MultiUser  NixMode = "MultiUser"
	SingleUser NixMode = "SingleUser"
)

func getNixExecutable() string {
	// 1. Check for user-defined path in config
	if nixBinPath, err := config.GetNixBinPath(); err == nil && nixBinPath != "" {
		if _, err := os.Stat(nixBinPath); err == nil {
			return nixBinPath
		}
	}

	// 2. Check if 'nix' is in the system's PATH
	if path, err := exec.LookPath("nix"); err == nil {
		return path
	}

	// 2. Check the multi-user installation path
	multiUserPath := "/nix/var/nix/profiles/default/bin/nix"
	if _, err := os.Stat(multiUserPath); err == nil {
		return multiUserPath
	}

	// 3. Check the single-user installation path
	home, err := os.UserHomeDir()
	if err == nil {
		singleUserPath := home + "/.local/state/nix/profiles/profile/bin/nix"
		if _, err := os.Stat(singleUserPath); err == nil {
			return singleUserPath
		}
	}

	return "" // Return empty if not found in any location
}

func RunCommand(command string, args ...string) (string, error) {
	if command == "nix" {
		command = getNixExecutable()
		if command == "" {
			return "", fmt.Errorf("nix executable not found: please ensure Nix is installed and in your PATH")
		}
	}
	if strings.HasSuffix(command, "nix") {
		args = append([]string{"--extra-experimental-features", "nix-command", "--extra-experimental-features", "flakes"}, args...)
	}
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command '%s %s': %s\n%s", command, strings.Join(args, " "), err, string(output))
	}
	return string(output), nil
}

func RunInteractiveCommand(command string, args ...string) error {
	if command == "nix" {
		command = getNixExecutable()
		if command == "" {
			return fmt.Errorf("nix executable not found: please ensure Nix is installed and in your PATH")
		}
	}
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RunSudoCommand(password string, args ...string) (string, error) {
	cmd := exec.Command("sudo", "-S")
	cmd.Stdin = strings.NewReader(password)
	cmd.Args = append(cmd.Args, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running sudo command: %s\n%s", err, string(output))
	}
	return string(output), nil
}

func RunCommandInDir(dir, command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command in dir '%s': %s\n%s", dir, err, string(output))
	}
	return string(output), nil
}

func GetNixMode() NixMode {
	// A simple way to check if we're on NixOS is to see if /etc/NIXOS is present.
	// This is not foolproof, but it's a good heuristic.
	// Check for NixOS by looking for the /etc/NIXOS file.
	if _, err := os.Stat("/etc/NIXOS"); err == nil {
		return NixOS
	}
	// Check for a multi-user installation by looking for the /nix/var/nix/profiles/per-user directory.
	if _, err := os.Stat("/nix/var/nix/profiles/per-user"); err == nil {
		return MultiUser
	}
	// Check for a single-user installation by looking for the ~/.nix-profile directory.
	home, err := os.UserHomeDir()
	if err == nil {
		if _, err := os.Stat(home + "/.nix-profile"); err == nil {
			return SingleUser
		}
	}
	return None
}

func UpdateFlake(path string) error {
	_, err := RunCommand("nix", "flake", "update", "--flake", path)
	return err
}

func CommitAndPush(path, message string) error {
	if _, err := RunCommandInDir(path, "git", "add", "."); err != nil {
		return err
	}
	if _, err := RunCommandInDir(path, "git", "commit", "-m", message); err != nil {
		return err
	}
	if _, err := RunCommandInDir(path, "git", "push"); err != nil {
		return err
	}
	return nil
}

// IsNixInstalled checks if the 'nix' command is available in the system's PATH.
func IsNixInstalled() bool {
	// First, check if 'nix' is in the system's PATH
	_, err := exec.LookPath("nix")
	if err == nil {
		return true
	}
	// If not in PATH, check the fallback location
	if _, err := os.Stat("/nix/var/nix/profiles/default/bin/nix"); err == nil {
		return true
	}
	return false
}

func RunCommandInNewTerminal(command string, args ...string) error {
	if command == "nix" {
		command = getNixExecutable()
		if command == "" {
			return fmt.Errorf("nix executable not found: please ensure Nix is installed and in your PATH")
		}
	}
	cmd := exec.Command(command, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func Commit(path, message string) error {
	if _, err := RunCommandInDir(path, "git", "add", "."); err != nil {
		return err
	}
	if _, err := RunCommandInDir(path, "git", "commit", "-m", message); err != nil {
		return err
	}
	return nil
}
