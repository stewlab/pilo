package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"os/user"
	"path/filepath"
	"pilo/internal/config"
	"pilo/internal/nix"
	"strings"

	"github.com/acarl005/stripansi"
)

// BaseConfig represents the structure of the base-config.json file.
type BaseConfig struct {
	System struct {
		Type string `json:"type"`
	} `json:"system"`
}

// getSystemType reads the system type from the base-config.json file.
func getSystemType() (string, error) {
	configPath := filepath.Join(config.GetFlakePath(), "base-config.json")
	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("could not read base-config.json: %w", err)
	}

	var baseConfig BaseConfig
	if err := json.Unmarshal(file, &baseConfig); err != nil {
		return "", fmt.Errorf("could not unmarshal base-config.json: %w", err)
	}

	return baseConfig.System.Type, nil
}

// RunCommandAndCommit executes a command and commits the changes if the command is a trigger.
func RunCommandAndCommit(commandName string, password string, args ...string) (string, error) {
	triggers, err := config.GetCommitTriggers()
	if err != nil {
		return "", fmt.Errorf("could not get commit triggers: %w", err)
	}

	commit := false
	for _, trigger := range triggers {
		if trigger == commandName {
			commit = true
			break
		}
	}

	var output string
	if password != "" {
		output, err = nix.RunSudoCommand(password, args...)
	} else {
		output, err = nix.RunCommand(args[0], args[1:]...)
	}

	if err != nil {
		return output, err
	}

	if commit {
		pushOnCommit, err := config.GetPushOnCommit()
		if err != nil {
			return output, fmt.Errorf("could not get push on commit setting: %w", err)
		}
		if pushOnCommit {
			if err := nix.CommitAndPush(config.GetInstallPath(), fmt.Sprintf("pilo: %s", commandName)); err != nil {
				return output, fmt.Errorf("could not commit and push changes: %w", err)
			}
		} else {
			if err := nix.Commit(config.GetInstallPath(), fmt.Sprintf("pilo: %s", commandName)); err != nil {
				return output, fmt.Errorf("could not commit changes: %w", err)
			}
		}
	}

	return output, nil
}

// Rebuild rebuilds the system configuration.
func Rebuild(flakePath, password, nixpkgsUrl, homeManagerUrl string) (string, error) {
	if flakePath == "" {
		flakePath = config.GetFlakePath()
	}

	// Use provided URLs, otherwise get from config
	if nixpkgsUrl == "" {
		nixpkgsUrl = config.GetNixpkgsUrl()
	}
	if homeManagerUrl == "" {
		homeManagerUrl = config.GetHomeManagerUrl()
	}

	var args []string
	var out string
	var err error
	switch nix.GetNixMode() {
	case nix.NixOS:
		fmt.Println("NixOS detected, running nixos-rebuild...")
		args = []string{"nixos-rebuild", "switch", "--flake", flakePath + "#nixos"}
		if nixpkgsUrl != "" {
			args = append(args, "--override-input", "nixpkgs", nixpkgsUrl)
		}
		if homeManagerUrl != "" {
			args = append(args, "--override-input", "home-manager", homeManagerUrl)
		}
		out, err = RunCommandAndCommit("rebuild", password, args...)
	case nix.MultiUser, nix.SingleUser:
		var u *user.User
		u, err = user.Current()
		if err != nil {
			return "", fmt.Errorf("could not get current user: %w", err)
		}
		username := u.Username

		systemType, err := getSystemType()
		if err != nil {
			return "", err
		}

		fmt.Println("Home Manager detected, running home-manager switch...")
		flakeRef := fmt.Sprintf("%s#%s@%s", flakePath, username, systemType)
		args = []string{"home-manager", "switch", "--flake", flakeRef}
		if nixpkgsUrl != "" {
			args = append(args, "--override-input", "nixpkgs", nixpkgsUrl)
		}
		if homeManagerUrl != "" {
			args = append(args, "--override-input", "home-manager", homeManagerUrl)
		}
		out, err = RunCommandAndCommit("rebuild", "", args...)
	default:
		err = fmt.Errorf("no supported Nix installation found")
	}
	if err != nil {
		return out, fmt.Errorf("rebuild failed: %w\nOutput:\n%s", err, out)
	}

	// Refresh the data
	if _, err := GetInstalledPackages(); err != nil {
		return out, fmt.Errorf("failed to refresh packages: %w", err)
	}
	if _, err := ListDevshells(); err != nil {
		return out, fmt.Errorf("failed to refresh devshells: %w", err)
	}
	if _, err := GetAliases(); err != nil {
		return out, fmt.Errorf("failed to refresh aliases: %w", err)
	}

	return out, err
}

// Update updates the flake inputs.
func Update(inputName string) (string, error) {
	fmt.Println("Updating flake inputs...")
	flakePath := config.GetFlakePath()
	fmt.Printf("DEBUG: Running 'nix flake update' on flake: %s\n", flakePath)
	if inputName != "" {
		fmt.Printf("Updating input: %s\n", inputName)
		return nix.RunCommand("nix", "flake", "lock", "--update-input", inputName, "--flake", flakePath)
	}
	return nix.RunCommand("nix", "flake", "update", "--flake", flakePath)
}

// InstallConfig installs the Nix configuration.
func InstallConfig(path, registry, password string) error {
	fmt.Println("Adding to Nix registry...")
	if _, err := nix.RunCommand("nix", "registry", "add", registry, path); err != nil {
		return err
	}

	// fmt.Println("Applying configuration...")
	// switch GetNixMode() {
	// case NixOS:
	// 	fmt.Println("NixOS detected, running nixos-rebuild...")
	// 	return RunSudoCommand(password, "nixos-rebuild", "switch", "--flake", "path:"+path+"#nixos")
	// case MultiUser, SingleUser:
	// 	currentUser, err := user.Current()
	// 	if err != nil {
	// 		return fmt.Errorf("could not get current user: %w", err)
	// 	}
	// 	username := currentUser.Username
	// 	fmt.Println("Home Manager detected, running home-manager switch...")
	// 	flakeRef := fmt.Sprintf("path:%s#%s@x86_64-linux", path, username)
	// 	return RunCommand("home-manager", "switch", "--flake", flakeRef)
	// default:
	// 	return fmt.Errorf("no supported Nix installation found")
	// }
	return nil

}

func Upgrade() (string, error) {
	fmt.Println("Upgrading...")
	return nix.RunCommand("nix", "flake", "update", "pilo")
}

func GC() (string, error) {
	fmt.Println("Running garbage collection...")
	return nix.RunCommand("nix", "store", "gc")
}

func Rollback(password string) (string, error) {
	var out string
	var err error
	switch nix.GetNixMode() {
	case nix.NixOS:
		fmt.Println("Rolling back NixOS to previous generation...")
		out, err = nix.RunSudoCommand(password, "nixos-rebuild", "switch", "--rollback")
	case nix.MultiUser, nix.SingleUser:
		fmt.Println("Rolling back Home Manager to previous generation...")
		out, err = nix.RunCommand("home-manager", "switch", "--rollback")
	default:
		err = fmt.Errorf("no supported Nix installation found for rollback")
	}
	if err != nil {
		return out, fmt.Errorf("rollback failed: %w\nOutput:\n%s", err, out)
	}
	return out, err
}

// ListGenerations lists the NixOS generations in reverse chronological order.
func ListGenerations() (string, error) {
	fmt.Println("Listing NixOS generations...")
	out, err := nix.RunCommand("nix", "profile", "history", "--profile", "/nix/var/nix/profiles/system")
	if err != nil {
		return "", err
	}
	sanitizedOutput := stripansi.Strip(string(out))
	lines := strings.Split(strings.TrimSpace(sanitizedOutput), "\n")
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	return strings.Join(lines, "\n"), nil
}

// InstallNix ensures that Nix is installed on the system.
func InstallNix() error {
	return EnsureNixInstalled()
}

// GetPendingActions returns a list of pending actions.
func GetPendingActions() ([]string, error) {
	flakePath := config.GetFlakePath()
	var actions []string
	if hasUncommittedChanges(filepath.Dir(flakePath)) {
		actions = append(actions, "Uncommitted changes")
	}
	unpushedCommits, err := getUnpushedCommits(filepath.Dir(flakePath))
	if err != nil {
		// Handle error, maybe log it
	}
	if unpushedCommits > 0 {
		actions = append(actions, fmt.Sprintf("%d unpushed commits", unpushedCommits))
	}
	return actions, nil
}

// hasUncommittedChanges checks if there are uncommitted changes in the git repository.
func hasUncommittedChanges(path string) bool {
	cmd := exec.Command("git", "-C", path, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		// If git status fails, assume no changes to be safe
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// getUnpushedCommits checks if there are unpushed commits in the git repository.
func getUnpushedCommits(path string) (int, error) {
	cmd := exec.Command("git", "-C", path, "rev-list", "--count", "@{u}..")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	count := 0
	fmt.Sscanf(string(output), "%d", &count)
	return count, nil
}
