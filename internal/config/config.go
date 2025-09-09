package config

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"

	"fyne.io/fyne/v2"
)

// BaseConfig defines the structure of the base-config.json file.
type System struct {
	Username string `json:"username"`
	Desktop  string `json:"desktop"`
	Type     string `json:"type"`
	Ollama   Ollama `json:"ollama"`
}

type Ollama struct {
	Models string `json:"models"`
}

type Package struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Installed   bool   `json:"installed"`
}

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
}

type BaseConfig struct {
	CommitTriggers []string          `json:"commit_triggers"`
	Packages       []Package         `json:"packages"`
	Aliases        map[string]string `json:"aliases"`
	PushOnCommit   bool              `json:"push_on_commit"`
	RemoteURL      string            `json:"remote_url"`
	RemoteBranch   string            `json:"remote_branch"`
	System         System            `json:"system"`
	Users          []User            `json:"users"`
	NixBinPath     string            `json:"nix_bin_path"`
}

var App fyne.App

// Init initializes the application preferences
func Init(a fyne.App) {
	App = a
}

func GetFlakePath() string {
	return filepath.Join(GetInstallPath(), "flake")
}

// GetInstallPath retrieves the installation path from preferences.
func GetInstallPath() string {
	if App == nil {
		// This is a fallback for CLI mode
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return filepath.Join(home, ".config", "pilo")
	}
	return App.Preferences().StringWithFallback("installationPath", must(os.UserHomeDir())+"/.config/pilo")
}

// readConfig reads and unmarshals the base-config.json file.
func ReadConfig() (*BaseConfig, error) {
	configPath := filepath.Join(GetInstallPath(), "flake", "base-config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file doesn't exist, create a default one and read it again.
			if err := writeDefaultConfig(configPath); err != nil {
				return nil, err
			}
			return ReadConfig()
		}
		return nil, err
	}

	var config BaseConfig
	if err := json.Unmarshal(data, &config); err != nil {
		// If unmarshalling fails, it might be due to an invalid config.
		// Create a default one and read it again.
		if err := writeDefaultConfig(configPath); err != nil {
			return nil, err
		}
		return ReadConfig()
	}
	return &config, nil
}

// writeDefaultConfig creates a default base-config.json file.
func writeDefaultConfig(path string) error {
	return GenerateBaseConfig(path)
}

// WriteConfig marshals and writes the config to base-config.json, ensuring slices are sorted.
func WriteConfig(config *BaseConfig) error {
	// Sort all string slices to ensure canonical representation
	sort.Strings(config.CommitTriggers)
	sort.Slice(config.Packages, func(i, j int) bool {
		return config.Packages[i].Name < config.Packages[j].Name
	})

	configPath := filepath.Join(GetInstallPath(), "flake", "base-config.json")
	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, newData, 0644)
}

// GetNixpkgsUrl retrieves the Nixpkgs URL from preferences.
func GetNixpkgsUrl() string {
	if App == nil {
		return "github:NixOS/nixpkgs/nixos-25.05"
	}
	return App.Preferences().StringWithFallback("nixpkgsUrl", "github:NixOS/nixpkgs/nixos-25.05")
}

// SetNixpkgsUrl sets the Nixpkgs URL in preferences.
func SetNixpkgsUrl(url string) {
	if App == nil {
		return
	}
	App.Preferences().SetString("nixpkgsUrl", url)
}

// GetHomeManagerUrl retrieves the Home Manager URL from preferences.
func GetHomeManagerUrl() string {
	if App == nil {
		return "github:nix-community/home-manager/release-25.05"
	}
	return App.Preferences().StringWithFallback("homeManagerUrl", "github:nix-community/home-manager/release-25.05")
}

// SetHomeManagerUrl sets the Home Manager URL in preferences.
func SetHomeManagerUrl(url string) {
	if App == nil {
		return
	}
	App.Preferences().SetString("homeManagerUrl", url)
}

func must(s string, e error) string {
	if e != nil {
		panic(e)
	}
	return s
}

// GetNixInstallCmd retrieves the Nix installation command from preferences.
func GetNixInstallCmd() string {
	if App == nil {
		return "curl --proto '=https' --tlsv1.2 -L https://nixos.org/nix/install | sh -s -- --no-daemon"
	}
	return App.Preferences().StringWithFallback("nixInstallCmd", "curl --proto '=https' --tlsv1.2 -L https://nixos.org/nix/install | sh -s -- --no-daemon")
}

// SetNixInstallCmd sets the Nix installation command in preferences.
func SetNixInstallCmd(cmd string) {
	if App == nil {
		return
	}
	App.Preferences().SetString("nixInstallCmd", cmd)
}

// GetLogHistoryRetention retrieves the log history retention from preferences.
func GetLogHistoryRetention() int {
	if App == nil {
		return 1000
	}
	return App.Preferences().IntWithFallback("logHistoryRetention", 1000)
}

// SetLogHistoryRetention sets the log history retention in preferences.
func SetLogHistoryRetention(retention int) {
	if App == nil {
		return
	}
	App.Preferences().SetInt("logHistoryRetention", retention)
}

// GetLogs retrieves the logs from preferences.
func GetLogs() []string {
	if App == nil {
		return []string{}
	}
	return App.Preferences().StringListWithFallback("logs", []string{})
}

// SetLogs sets the logs in preferences.
func SetLogs(logs []string) {
	if App == nil {
		return
	}
	App.Preferences().SetStringList("logs", logs)
}

// AddLogEntry adds a new log entry with a timestamp.
func AddLogEntry(event string) {
	if App == nil {
		return
	}
	logs := GetLogs()
	timestamp := App.Preferences().StringWithFallback("currentTime", "unknown time") // Use current time from preferences or fallback
	newEntry := fmt.Sprintf("[%s] %s", timestamp, event)

	// Prepend the new entry to the slice
	logs = append([]string{newEntry}, logs...)

	// Trim logs to retention policy
	retention := GetLogHistoryRetention()
	if len(logs) > retention {
		logs = logs[:retention]
	}
	SetLogs(logs)
}

// GetCommitTriggers retrieves the commit triggers from the base config file.
func GetCommitTriggers() ([]string, error) {
	config, err := ReadConfig()
	if err != nil {
		return nil, err
	}
	return config.CommitTriggers, nil
}

// SetCommitTriggers sets the commit triggers in the base config file.
func SetCommitTriggers(triggers []string) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}
	config.CommitTriggers = triggers
	return WriteConfig(config)
}

// GetRemoteUrl retrieves the remote URL from the base config file.
func GetRemoteUrl() (string, error) {
	config, err := ReadConfig()
	if err != nil {
		return "", err
	}
	return config.RemoteURL, nil
}

// SetRemoteUrl sets the remote URL in the base config file.
func SetRemoteUrl(url string) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}
	config.RemoteURL = url
	return WriteConfig(config)
}

// GetPushOnCommit retrieves the push on commit flag from the base config file.
func GetPushOnCommit() (bool, error) {
	config, err := ReadConfig()
	if err != nil {
		return false, err
	}
	return config.PushOnCommit, nil
}

// SetPushOnCommit sets the push on commit flag in the base config file.
func SetPushOnCommit(push bool) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}
	config.PushOnCommit = push
	return WriteConfig(config)
}

// GetRemoteBranch retrieves the remote branch from the base config file.
func GetRemoteBranch() (string, error) {
	config, err := ReadConfig()
	if err != nil {
		return "main", err
	}
	if config.RemoteBranch == "" {
		return "main", nil
	}
	return config.RemoteBranch, nil
}

// SetRemoteBranch sets the remote branch in the base config file.
func SetRemoteBranch(branch string) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}
	config.RemoteBranch = branch
	return WriteConfig(config)
}

// GetSystem retrieves the system from the base config file.
func GetSystem() (System, error) {
	config, err := ReadConfig()
	if err != nil {
		return System{}, err
	}
	return config.System, nil
}

// SetSystem sets the system in the base config file.
func SetSystem(system System) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}
	config.System = system
	return WriteConfig(config)
}

// GetUsers retrieves the list of users from the base config file.
func GetUsers() ([]User, error) {
	config, err := ReadConfig()
	if err != nil {
		return nil, err
	}
	return config.Users, nil
}

// SetUsers sets the list of users in the base config file.
func SetUsers(users []User) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}
	config.Users = users
	return WriteConfig(config)
}

// GetUsername retrieves the username from the base config file, or the system username if not set.
func GetUsername() (string, error) {
	config, err := ReadConfig()
	if err != nil {
		// If config file doesn't exist, create a default one and retry.
		if os.IsNotExist(err) {
			if err := writeDefaultConfig(filepath.Join(GetInstallPath(), "flake", "base-config.json")); err != nil {
				return "", err
			}
			config, err = ReadConfig()
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	// If username is empty in config, use system username and update the config.
	if config.System.Username == "" {
		currentUser, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("could not get current user: %w", err)
		}
		config.System.Username = currentUser.Username
		if err := WriteConfig(config); err != nil {
			// Log or handle error, but proceed with the username.
		}
	}

	return config.System.Username, nil
}

// SetUsername sets the username in the base config file.
func SetUsername(username string) error {
	config, err := ReadConfig()
	// If the file doesn't exist, we create a new config
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if config == nil {
		config = &BaseConfig{}
	}

	config.System.Username = username
	return WriteConfig(config)
}

// GetType retrieves the system type from the base config file.
func GetType() (string, error) {
	config, err := ReadConfig()
	if err != nil {
		return "", err
	}
	if config.System.Type == "" {
		return "x86_64-linux", nil
	}
	return config.System.Type, nil
}

// SetType sets the system type in the base config file.
func SetType(systemType string) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}
	config.System.Type = systemType
	return WriteConfig(config)
}

// // GetUsers retrieves the list of users from the base config file.
// func GetUsers() ([]User, error) {
// 	config, err := ReadConfig()
// 	if err != nil {
// 		return nil, err
// 	}
// 	return config.Users, nil
// }

// // SetUsers sets the list of users in the base config file.
// func SetUsers(users []User) error {
// 	config, err := ReadConfig()
// 	if err != nil {
// 		return err
// 	}
// 	config.Users = users
// 	return WriteConfig(config)
// }

// GetDesktop retrieves
func GetDesktop() (string, error) {
	config, err := ReadConfig()
	if err != nil {
		return "", err
	}
	if config.System.Desktop == "" {
		return "plasma", nil
	}
	return config.System.Desktop, nil
}

// SetDesktop sets the desktop used in NixOS
func SetDesktop(desktop string) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}
	config.System.Desktop = desktop
	return WriteConfig(config)
}

// GetInstallFromRemote retrieves the install from remote flag from preferences.
func GetInstallFromRemote() bool {
	if App == nil {
		return false
	}
	return App.Preferences().BoolWithFallback("installFromRemote", false)
}

// GetCustomTerminal retrieves the custom terminal command from preferences.
func GetCustomTerminal() string {
	if App == nil {
		return ""
	}
	return App.Preferences().StringWithFallback("customTerminal", "")
}

// SetCustomTerminal sets the custom terminal command in preferences.
func SetCustomTerminal(cmd string) {
	if App == nil {
		return
	}
	App.Preferences().SetString("customTerminal", cmd)
}

// GetRemoteGitUrl retrieves the remote Git URL from preferences.
func GetRemoteGitUrl() string {
	if App == nil {
		return ""
	}
	return App.Preferences().StringWithFallback("remoteGitUrl", "")
}

// SetRemoteGitUrl sets the remote Git URL in preferences.
func SetRemoteGitUrl(url string) {
	if App == nil {
		return
	}
	App.Preferences().SetString("remoteGitUrl", url)
}

// GetSshKeyPath retrieves the SSH key path from preferences.
func GetSshKeyPath() string {
	if App == nil {
		return ""
	}
	return App.Preferences().StringWithFallback("sshKeyPath", "")
}

// SetSshKeyPath sets the SSH key path in preferences.
func SetSshKeyPath(path string) {
	if App == nil {
		return
	}
	App.Preferences().SetString("sshKeyPath", path)
}

// GetRegistryName retrieves the registry name from preferences.
func GetRegistryName() string {
	if App == nil {
		return "pilo"
	}
	return App.Preferences().StringWithFallback("registryName", "pilo")
}

// SetRegistryName sets the registry name in preferences.
func SetRegistryName(name string) {
	if App == nil {
		return
	}
	App.Preferences().SetString("registryName", name)
}

// GetNixBinPath retrieves the Nix binary path from the base config file.
func GetNixBinPath() (string, error) {
	config, err := ReadConfig()
	if err != nil {
		return "", err
	}
	return config.NixBinPath, nil
}

// SetNixBinPath sets the Nix binary path in the base config file.
func SetNixBinPath(path string) error {
	config, err := ReadConfig()
	if err != nil {
		return err
	}
	config.NixBinPath = path
	return WriteConfig(config)
}
