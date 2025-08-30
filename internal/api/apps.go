package api

import (
	"fmt"
	"os"
	"path/filepath"
	"pilo/internal/config"
	"text/template"
)

func getPackagesDir() string {
	return filepath.Join(config.GetFlakePath(), "packages")
}

const (
	appTmpl = `{ pkgs, ... }:

pkgs.stdenv.mkDerivation {
  pname = "{{ .Pname }}";
  version = "{{ .Version }}";

  src = pkgs.fetchurl {
    url = "{{ .URL }}";
    sha256 = "{{ .Sha256 }}";
  };

  nativeBuildInputs = [ pkgs.nodejs ];

  installPhase = ''
    npm install --ignore-scripts
    mkdir -p $out/bin
    ln -s $PWD/node_modules/.bin/{{ .Pname }} $out/bin/{{ .Pname }}
  '';
}
`
)

// App represents a Flake application.
type App struct {
	Pname   string
	Version string
	URL     string
	Sha256  string
}

// AddApp creates a new Flake App file.
func AddApp(app App) error {
	tmpl, err := template.New("app").Parse(appTmpl)
	if err != nil {
		return fmt.Errorf("failed to parse app template: %w", err)
	}

	filePath := filepath.Join(getPackagesDir(), app.Pname+".nix")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create app file: %w", err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, app); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	path := config.GetInstallPath()
	if err := GitAdd(path); err != nil {
		return fmt.Errorf("could not add changes: %w", err)
	}

	return nil
}

// AddAppFromContent creates a new Flake App file from content.
func AddAppFromContent(pname, content string) error {
	filePath := filepath.Join(getPackagesDir(), pname+".nix")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create app file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to app file: %w", err)
	}

	path := config.GetInstallPath()
	if err := GitAdd(path); err != nil {
		return fmt.Errorf("could not add changes: %w", err)
	}

	return nil
}

// RemoveApp removes a Flake App file.
func RemoveApp(pname string) error {
	filePath := filepath.Join(getPackagesDir(), pname+".nix")
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove app file: %w", err)
	}
	path := config.GetInstallPath()
	if err := GitAdd(path); err != nil {
		return fmt.Errorf("could not add changes: %w", err)
	}

	return nil
}

// DuplicateApp duplicates a custom package file.
func DuplicateApp(pname string) error {
	originalPath := filepath.Join(getPackagesDir(), pname+".nix")
	content, err := os.ReadFile(originalPath)
	if err != nil {
		return err
	}

	// Find a new name
	i := 1
	var newName string
	for {
		newName = fmt.Sprintf("%s-%d", pname, i)
		newPath := filepath.Join(getPackagesDir(), newName+".nix")
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			break
		}
		i++
	}

	newPath := filepath.Join(getPackagesDir(), newName+".nix")
	return os.WriteFile(newPath, content, 0644)
}

// RenameApp renames a custom package file.
func RenameApp(oldName, newName string) error {
	oldPath := filepath.Join(getPackagesDir(), oldName+".nix")
	newPath := filepath.Join(getPackagesDir(), newName+".nix")
	return os.Rename(oldPath, newPath)
}

// GetAppTemplate returns the raw template for a new app.
func GetAppTemplate() string {
	return appTmpl
}

// ListApps lists all the custom applications.
func ListApps() ([]string, error) {
	files, err := os.ReadDir(getPackagesDir())
	if err != nil {
		return nil, fmt.Errorf("failed to read packages dir: %w", err)
	}

	var apps []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".nix" {
			apps = append(apps, file.Name())
		}
	}

	return apps, nil
}

// GetAppContent returns the content of a specific app file.
func GetAppContent(pname string) (string, error) {
	filePath := filepath.Join(getPackagesDir(), pname+".nix")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read app file: %w", err)
	}
	return string(content), nil
}

// UpdateApp updates the content of a specific app file.
func UpdateApp(pname, content string) error {
	filePath := filepath.Join(getPackagesDir(), pname+".nix")
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write to app file: %w", err)
	}
	return nil
}
