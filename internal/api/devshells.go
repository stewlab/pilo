package api

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"pilo/internal/config"
	"pilo/internal/nix"
	"strings"
)

func getDevshellTemplatesDir() string {
	return filepath.Join(config.GetFlakePath(), "devshells", "templates")
}

// Devshell represents a development shell template.
type Devshell struct {
	Name        string
	Type        string // "Normal" or "FHS"
	Description string
}

// InitDevshellFromTemplate creates a new devshell project from a template.
func InitDevshellFromTemplate(templateName, targetDir string) error {
	templatePath := filepath.Join(getDevshellTemplatesDir(), templateName, "flake.nix")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("devshell template '%s' not found", templateName)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("could not create target directory '%s': %w", targetDir, err)
	}

	flakeNixPath := filepath.Join(targetDir, "flake.nix")
	if _, err := os.Stat(flakeNixPath); err == nil {
		return fmt.Errorf("a 'flake.nix' file already exists in '%s'", targetDir)
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("could not read template file: %w", err)
	}

	return os.WriteFile(flakeNixPath, content, 0644)
}

// ListDevshells lists all available devshell templates.
func ListDevshells() ([]Devshell, error) {
	entries, err := os.ReadDir(getDevshellTemplatesDir())
	if err != nil {
		return nil, err
	}

	var devshells []Devshell
	for _, entry := range entries {
		if entry.IsDir() {
			templateName := entry.Name()
			flakePath := filepath.Join(getDevshellTemplatesDir(), templateName, "flake.nix")
			if _, err := os.Stat(flakePath); err == nil {
				description := "A development shell template."
				shellType := "Normal"
				if strings.Contains(templateName, "fhs") {
					description = "An FHS development shell template."
					shellType = "FHS"
				}
				devshells = append(devshells, Devshell{Name: templateName, Type: shellType, Description: description})
			}
		}
	}
	return devshells, nil
}

// EnterDevshell starts a new terminal in the specified devshell directory.
func EnterDevshell(directory string) error {
	terminalCmd := config.GetCustomTerminal()
	if terminalCmd == "" {
		terminalCmd = "xterm"
	}
	// The command needs to change to the directory first, then run nix develop.
	fullCmd := fmt.Sprintf("cd %s && nix develop", directory)
	args := []string{"-e", "sh", "-c", fullCmd}
	return nix.RunCommandInNewTerminal(terminalCmd, args...)
}

// RunInDevshell runs a command in the specified devshell directory.
func RunInDevshell(directory, command string) (string, error) {
	cmd := exec.Command("nix", "develop", directory, "--command", "sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command in devshell: %w\nOutput: %s", err, string(output))
	}
	return string(output), nil
}

// Develop enters a persistent development shell.
func Develop(args []string) error {
	fmt.Println("Entering a development shell...")
	flakePath := config.GetFlakePath()
	// Default to the 'default' shell if no arguments are provided
	shell := "default"
	if len(args) > 0 {
		shell = args[0]
	}

	// Construct the flake reference
	flakeRef := fmt.Sprintf("%s#%s", flakePath, shell)

	// The command should be "nix", "develop", "<flakeRef>"
	return nix.RunInteractiveCommand("nix", "develop", flakeRef)
}
