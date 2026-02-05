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

// GetDevshellContent returns the content of a devshell's flake.nix file.
func GetDevshellContent(name string) (string, error) {
	userDevshellsDir := config.GetUserDevshellsDir()
	flakeNixPath := filepath.Join(userDevshellsDir, name, "flake.nix")
	content, err := os.ReadFile(flakeNixPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// UpdateDevshell updates the content of a devshell's flake.nix file.
func UpdateDevshell(name, content string) error {
	userDevshellsDir := config.GetUserDevshellsDir()
	flakeNixPath := filepath.Join(userDevshellsDir, name, "flake.nix")
	return os.WriteFile(flakeNixPath, []byte(content), 0644)
}

// RenameDevShell renames a devshell directory.
func RenameDevShell(oldName, newName string) error {
	userDevshellsDir := config.GetUserDevshellsDir()
	oldPath := filepath.Join(userDevshellsDir, oldName)
	newPath := filepath.Join(userDevshellsDir, newName)
	return os.Rename(oldPath, newPath)
}

func getDevshellTemplatesDir() string {
	return filepath.Join(config.GetFlakePath(), "devshells", "templates")
}

// Devshell represents a development shell template.
type Devshell struct {
	Name        string
	Path        string // Absolute path to the devshell directory
	Type        string // "Normal" or "FHS"
	Description string
}

// InitDevshellFromTemplate creates a new devshell project from a template.
func InitDevshellFromTemplate(templateName, targetDir string) error {
	templatePath := filepath.Join(getDevshellTemplatesDir(), templateName, "flake.nix")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("devshell template '%s' not found", templateName)
	}

	// Ensure user devshells directory exists
	userDevshellsDir := config.GetUserDevshellsDir()
	if err := os.MkdirAll(userDevshellsDir, 0755); err != nil {
		return fmt.Errorf("could not create user devshells directory '%s': %w", userDevshellsDir, err)
	}

	// Prevent absolute or parent directory traversal
	cleanTargetDir := filepath.Clean(targetDir)
	if filepath.IsAbs(cleanTargetDir) || strings.HasPrefix(cleanTargetDir, "..") || cleanTargetDir == "" {
		return fmt.Errorf("invalid target directory name: '%s'", targetDir)
	}

	absTargetDir := filepath.Join(userDevshellsDir, cleanTargetDir)
	if err := os.MkdirAll(absTargetDir, 0755); err != nil {
		return fmt.Errorf("could not create target directory '%s': %w", absTargetDir, err)
	}

	flakeNixPath := filepath.Join(absTargetDir, "flake.nix")
	if _, err := os.Stat(flakeNixPath); err == nil {
		return fmt.Errorf("a 'flake.nix' file already exists in '%s'", absTargetDir)
	}

	content, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("could not read template file: %w", err)
	}

	return os.WriteFile(flakeNixPath, content, 0644)
}

// ListDevshellTemplates lists all available devshell templates.
func ListDevshellTemplates() ([]Devshell, error) {
	templatesDir := getDevshellTemplatesDir()
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		// If the templates directory doesn't exist, it means no templates are available.
		// This is not an error, just an empty list.
		return []Devshell{}, nil
	} else if err != nil {
		return nil, fmt.Errorf("failed to access devshell templates directory: %w", err)
	}

	entries, err := os.ReadDir(templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read devshell templates directory: %w", err)
	}

	var devshells []Devshell
	for _, entry := range entries {
		if entry.IsDir() {
			templateName := entry.Name()
			templatePath := filepath.Join(getDevshellTemplatesDir(), templateName)
			flakePath := filepath.Join(templatePath, "flake.nix")
			if _, err := os.Stat(flakePath); err == nil {
				description := "A development shell template."
				shellType := "Normal"
				if strings.Contains(templateName, "fhs") {
					description = "An FHS development shell template."
					shellType = "FHS"
				}
				devshells = append(devshells, Devshell{Name: templateName, Path: templatePath, Type: shellType, Description: description})
			}
		}
	}
	return devshells, nil
}

// ListUserDevshells lists all user-created devshells.
func ListUserDevshells() ([]Devshell, error) {
	// For now, let's assume user devshells are in a specific directory.
	// We need to define where user devshells are stored.
	// For this example, let's assume they are in config.GetUserDevshellsDir()
	// If this function doesn't exist, we'll need to create it or define a different path.
	userDevshellsDir := config.GetUserDevshellsDir()
	if _, err := os.Stat(userDevshellsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(userDevshellsDir, 0755); err != nil {
			return nil, fmt.Errorf("could not create user devshells directory '%s': %w", userDevshellsDir, err)
		}
		return []Devshell{}, nil // Directory created, but no devshells yet
	} else if err != nil {
		return nil, fmt.Errorf("failed to access user devshells directory: %w", err)
	}

	entries, err := os.ReadDir(userDevshellsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read user devshells directory: %w", err)
	}

	var devshells []Devshell
	for _, entry := range entries {
		if entry.IsDir() {
			devshellName := entry.Name()
			devshellPath := filepath.Join(userDevshellsDir, devshellName)
			flakePath := filepath.Join(devshellPath, "flake.nix")
			if _, err := os.Stat(flakePath); err == nil {
				// We can try to read the flake.nix to get more details if needed,
				// but for now, just use the directory name as the devshell name.
				devshells = append(devshells, Devshell{Name: devshellName, Path: devshellPath, Type: "User", Description: "User-created devshell"})
			}
		}
	}
	return devshells, nil
}

// EditDevshell opens the devshell's flake.nix in the default editor.
func EditDevshell(devshellPath string) error {
	flakeNixPath := filepath.Join(devshellPath, "flake.nix")
	editorCmd := os.Getenv("EDITOR")
	if editorCmd == "" {
		editorCmd = "vi" // Default to vi if EDITOR is not set
	}
	cmd := exec.Command(editorCmd, flakeNixPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// DuplicateDevshell duplicates an existing devshell.
func DuplicateDevshell(sourcePath, newName string) error {
	userDevshellsDir := config.GetUserDevshellsDir()
	targetPath := filepath.Join(userDevshellsDir, newName)

	if _, err := os.Stat(targetPath); err == nil {
		return fmt.Errorf("devshell with name '%s' already exists", newName)
	}

	// Copy the entire directory
	cmd := exec.Command("cp", "-r", sourcePath, targetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to duplicate devshell: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// DeleteDevshell deletes a user-created devshell.
func DeleteDevshell(devshellPath string) error {
	return os.RemoveAll(devshellPath)
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
