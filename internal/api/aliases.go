package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"pilo/internal/config"
)

func getAliasesFile() string {
	return filepath.Join(config.GetFlakePath(), "aliases.json")
}

// Alias represents a custom command alias.
type Alias struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// GetAliases reads the aliases from the JSON file.
func GetAliases() (map[string]string, error) {
	data, err := os.ReadFile(getAliasesFile())
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("error reading aliases file: %w", err)
	}

	var aliases map[string]string
	if err := json.Unmarshal(data, &aliases); err != nil {
		return nil, fmt.Errorf("error unmarshaling aliases: %w", err)
	}

	return aliases, nil
}

// AddAlias adds a new alias to the JSON file.
func AddAlias(name, command string) error {
	aliases, err := GetAliases()
	if err != nil {
		return err
	}

	aliases[name] = command

	return saveAliases(aliases)
}

// RemoveAlias removes an alias from the JSON file.
func RemoveAlias(name string) error {
	aliases, err := GetAliases()
	if err != nil {
		return err
	}

	delete(aliases, name)

	return saveAliases(aliases)
}

// DuplicateAlias duplicates an alias.
func DuplicateAlias(name, command string) error {
	aliases, err := GetAliases()
	if err != nil {
		return err
	}

	// Find a new name
	i := 1
	var newName string
	for {
		newName = fmt.Sprintf("%s-%d", name, i)
		if _, exists := aliases[newName]; !exists {
			break
		}
		i++
	}

	aliases[newName] = command
	return saveAliases(aliases)
}

// UpdateAlias updates an existing alias. If the oldName is different from
// newName, it removes the old one.
func UpdateAlias(oldName, newName, command string) error {
	aliases, err := GetAliases()
	if err != nil {
		return err
	}

	if oldName != newName {
		delete(aliases, oldName)
	}
	aliases[newName] = command

	return saveAliases(aliases)
}

// saveAliases writes the aliases to the JSON file.
func saveAliases(aliases map[string]string) error {
	data, err := json.MarshalIndent(aliases, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling aliases: %w", err)
	}

	return os.WriteFile(getAliasesFile(), data, 0644)
}
