package nix

import (
	"fmt"
	"os"
	"os/exec"
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

func RunCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command '%s %s': %s\n%s", command, strings.Join(args, " "), err, string(output))
	}
	return string(output), nil
}

func RunInteractiveCommand(command string, args ...string) error {
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
	if _, err := RunCommand("stat", "/etc/NIXOS"); err == nil {
		return NixOS
	}
	return NonNixOS
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

func RunCommandInNewTerminal(command string, args ...string) error {
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
