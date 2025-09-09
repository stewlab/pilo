package config

import (
	"encoding/json"
	"os"
	"sort"
)

// GenerateBaseConfig creates a default BaseConfig and marshals it to JSON,
// then writes it to the specified file path.
func GenerateBaseConfig(filePath string) error {
	config := BaseConfig{
		Aliases: map[string]string{
			"..": "cd ..",
			"gs": "git status",
			"ll": "ls -alF",
		},
		CommitTriggers: []string{},
		Packages: []Package{
			{Name: "curl", Installed: true},
			{Name: "git", Installed: true},
			{Name: "pciutils", Installed: true},
			{Name: "wget", Installed: true},
			{Name: "blender", Installed: true},
			{Name: "delve", Installed: true},
			{Name: "gcc", Installed: true},
			{Name: "go", Installed: true},
			{Name: "godot", Installed: true},
			{Name: "gopls", Installed: true},
			{Name: "neovim", Installed: true},
			{Name: "pkg-config", Installed: true},
			{Name: "python3", Installed: true},
			{Name: "freeciv", Installed: true},
			{Name: "gzdoom", Installed: true},
			{Name: "minetest", Installed: true},
			{Name: "nethack", Installed: true},
			{Name: "nethack-x11", Installed: true},
			{Name: "superTuxKart", Installed: true},
			{Name: "wesnoth", Installed: true},
			{Name: "gimp", Installed: true},
			{Name: "inkscape", Installed: true},
			{Name: "chromium", Installed: true},
			{Name: "firefox", Installed: true},
			{Name: "mpv", Installed: true},
			{Name: "rclone", Installed: true},
			{Name: "vlc", Installed: true},
			{Name: "vscodium", Installed: true},
			{Name: "wine", Installed: true},
		},
		PushOnCommit: true,
		RemoteBranch: "main",
		RemoteURL:    "",
		System: System{
			Username: "",
			Desktop:  "",
			Type:     "x86_64-linux",
			Ollama: Ollama{
				Models: "",
			},
		},
		Users: []User{
			{
				Username: "nixuser",
				Email:    "nixuser@pilo",
				Name:     "Nix User",
			},
		},
		NixBinPath: "",
	}

	// Sort slices to ensure canonical representation
	sort.Strings(config.CommitTriggers)
	sort.Slice(config.Packages, func(i, j int) bool {
		return config.Packages[i].Name < config.Packages[j].Name
	})
	sort.Slice(config.Users, func(i, j int) bool {
		return config.Users[i].Username < config.Users[j].Username
	})

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
