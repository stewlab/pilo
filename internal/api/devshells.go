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

func getDevshellsDir() string {
	return filepath.Join(config.GetFlakePath(), "devshells")
}

// Devshell represents a development shell.
type Devshell struct {
	Name        string
	Type        string // "Normal" or "FHS"
	Description string
}

// AddDevshellWithContent creates a new devshell file with the given content.
// If the name is empty, a unique name is generated.
func AddDevshellWithContent(name, content string) error {
	if name == "" {
		// Find an unused name
		i := 1
		for {
			name = fmt.Sprintf("devshell-%d", i)
			filePath := filepath.Join(getDevshellsDir(), name+".nix")
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				break
			}
			i++
		}
	}

	filePath := filepath.Join(getDevshellsDir(), name+".nix")
	return os.WriteFile(filePath, []byte(content), 0644)
}

// RemoveDevshell removes a devshell file.
func RemoveDevshell(name string) error {
	filePath := filepath.Join(getDevshellsDir(), name+".nix")
	return os.Remove(filePath)
}

// DuplicateDevShell duplicates a devshell file.
func DuplicateDevShell(name string) error {
	originalPath := filepath.Join(getDevshellsDir(), name+".nix")
	content, err := os.ReadFile(originalPath)
	if err != nil {
		return err
	}

	// Find a new name
	i := 1
	var newName string
	for {
		newName = fmt.Sprintf("%s-%d", name, i)
		newPath := filepath.Join(getDevshellsDir(), newName+".nix")
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			break
		}
		i++
	}

	newPath := filepath.Join(getDevshellsDir(), newName+".nix")
	return os.WriteFile(newPath, content, 0644)
}

// RenameDevShell renames a devshell file.
func RenameDevShell(oldName, newName string) error {
	oldPath := filepath.Join(getDevshellsDir(), oldName+".nix")
	newPath := filepath.Join(getDevshellsDir(), newName+".nix")
	return os.Rename(oldPath, newPath)
}

// ListDevshells lists all available devshells.
func ListDevshells() ([]Devshell, error) {
	files, err := os.ReadDir(getDevshellsDir())
	if err != nil {
		return nil, err
	}

	var devshells []Devshell
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".nix") {
			name := strings.TrimSuffix(file.Name(), ".nix")
			// For simplicity, we'll just mark them as "Normal" for now.
			// A more robust solution would be to parse the file content.
			// A more robust solution would be to parse the file content.
			description := "A development shell."
			if strings.Contains(name, "fhs") {
				description = "An FHS development shell."
			}
			devshells = append(devshells, Devshell{Name: name, Type: "Normal", Description: description})
		}
	}
	return devshells, nil
}

// EnterDevshell starts a new terminal in the specified devshell.
func EnterDevshell(name, flakePath string) error {
	terminalCmd := config.GetCustomTerminal()
	if terminalCmd == "" {
		terminalCmd = "xterm"
	}
	return EnterDevshellWithTerminal(name, flakePath, terminalCmd)
}

// EnterDevshellWithTerminal starts a new terminal in the specified devshell with a custom terminal.
func EnterDevshellWithTerminal(name, flakePath, terminal string) error {
	if flakePath == "" {
		flakePath = config.GetFlakePath()
	}
	args := []string{"-e", "nix", "develop", flakePath + "#" + name}
	return nix.RunCommandInNewTerminal(terminal, args...)
}

// RunInDevshell runs a command in the specified devshell.
func RunInDevshell(name, command, flakePath string) (string, error) {
	if flakePath == "" {
		flakePath = config.GetFlakePath()
	}
	cmd := exec.Command("nix", "develop", flakePath+"#"+name, "--command", "sh", "-c", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("error running command in devshell: %w\nOutput: %s", err, string(output))
	}
	return string(output), nil
}

// GetDevshellContent returns the content of a devshell file.
func GetDevshellContent(name string) (string, error) {
	filePath := filepath.Join(getDevshellsDir(), name+".nix")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// UpdateDevshell updates the content of a devshell file.
func UpdateDevshell(name, content string) error {
	filePath := filepath.Join(getDevshellsDir(), name+".nix")
	return os.WriteFile(filePath, []byte(content), 0644)
}
