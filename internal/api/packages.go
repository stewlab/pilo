package api

import (
	"encoding/json"
	"fmt"
	"pilo/internal/config"
	"pilo/internal/nix"
	"sort"
	"strings"
)

func GetInstalledPackages() ([]config.Package, error) {
	return config.ReadPackagesConfig()
}

func AddPackage(packageName string) error {
	packages, err := config.ReadPackagesConfig()
	if err != nil {
		return err
	}

	for _, pkg := range packages {
		if pkg.Name == packageName {
			return fmt.Errorf("package '%s' already exists", packageName)
		}
	}

	newPackage := config.Package{Name: packageName, Installed: true}
	packages = append(packages, newPackage)

	if err := config.WritePackagesConfig(packages); err != nil {
		return err
	}

	return commitChanges(fmt.Sprintf("pilo: add package %s", packageName))
}

func RemovePackage(packageName string) error {
	packages, err := config.ReadPackagesConfig()
	if err != nil {
		return err
	}

	var newPackages []config.Package
	found := false
	for _, pkg := range packages {
		if pkg.Name == packageName {
			found = true
		} else {
			newPackages = append(newPackages, pkg)
		}
	}

	if !found {
		return fmt.Errorf("package '%s' not found", packageName)
	}

	if err := config.WritePackagesConfig(newPackages); err != nil {
		return err
	}

	return commitChanges(fmt.Sprintf("pilo: remove package %s", packageName))
}

func AddGitPackage(url string) error {
	// If the URL is a full GitHub URL, convert it to the github:owner/repo format
	if strings.HasPrefix(url, "https://github.com/") {
		parts := strings.Split(strings.TrimSuffix(url, ".git"), "/")
		if len(parts) >= 5 {
			url = fmt.Sprintf("github:%s/%s", parts[3], parts[4])
		}
	}
	return AddPackage(url)
}

// Search searches for packages in nixpkgs.
func Search(query []string, sortByPopularity bool, freeOnly bool, system string) ([]config.Package, error) {
	searchArgs := []string{"search", "nixpkgs", "--json"}
	searchArgs = append(searchArgs, query...)
	if system != "" {
		// pass system to nix search if provided (e.g., x86_64-linux)
		searchArgs = append(searchArgs, "--system", system)
	}
	out, err := nix.RunCommand("nix", searchArgs...)
	if err != nil {
		return nil, fmt.Errorf("error searching for packages: %w", err)
	}

	// Find the start of the JSON output
	jsonStart := strings.Index(out, "{")
	if jsonStart == -1 {
		return nil, fmt.Errorf("no JSON output found from search")
	}
	jsonOut := out[jsonStart:]

	var results map[string]struct {
		Pname       string `json:"pname"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(jsonOut), &results); err != nil {
		return nil, fmt.Errorf("error unmarshaling search results: %w", err)
	}

	// results map keys are attribute paths. We'll build a list of attributes
	// keyed by pname so we can show both the canonical pname and its
	// attribute path in results. This preserves attribute-level information
	// for the CLI/GUI while still allowing deduplication when desired.
	attrByPname := make(map[string][]string)
	descByAttr := make(map[string]string)
	for attr, result := range results {
		descByAttr[attr] = result.Description
		attrByPname[result.Pname] = append(attrByPname[result.Pname], attr)
	}

	// Sort pnames for deterministic output
	pnames := make([]string, 0, len(attrByPname))
	for pname := range attrByPname {
		pnames = append(pnames, pname)
	}
	sort.Strings(pnames)

	packages := make([]config.Package, 0, len(pnames))
	for _, pname := range pnames {
		// pick the first attribute for description display (deterministic)
		attrs := attrByPname[pname]
		sort.Strings(attrs)
		primaryAttr := attrs[0]
		packages = append(packages, config.Package{
			Name:        pname,
			Description: descByAttr[primaryAttr],
			Attribute:   primaryAttr,
		})
	}

	return packages, nil
}

func TempInstallPackage(packageName string) error {
	return nix.RunInteractiveCommand("nix-shell", "-p", packageName)
}

// Install installs packages.
func Install(packages []string) error {
	for _, pkg := range packages {
		if strings.HasPrefix(pkg, "github:") {
			if err := AddGitPackage(pkg); err != nil {
				return err
			}
			// After adding the git package, we need to update the flake
			if err := nix.UpdateFlake(config.GetInstallPath()); err != nil {
				return fmt.Errorf("failed to update flake: %w", err)
			}
			fmt.Printf("Added git package %s. Run 'nix flake update' and rebuild your system.", pkg)
			return nil
		}
	}

	if nix.GetNixMode() == nix.NixOS {
		fmt.Println("On NixOS, permanent packages should be added to configuration.nix.")
		fmt.Println("Providing a temporary shell with the requested packages...")
		args := append([]string{"-p"}, packages...)
		return nix.RunInteractiveCommand("nix-shell", args...)
	}
	fmt.Println("Installing packages with nix profile...")
	args := append([]string{"profile", "install"}, packages...)
	_, err := nix.RunCommand("nix", args...)
	return err
}

func List() (string, error) {
	fmt.Println("Listing user installed packages...")
	out, err := nix.RunCommand("nix", "profile", "list")
	if err != nil {
		return "", err
	}

	output := out
	if nix.GetNixMode() == nix.NixOS {
		output += "\n\nSystem packages on NixOS are managed declaratively in your configuration."
	}
	return output, nil
}

// Shell enters a temporary shell with the specified packages.
func Shell(packages []string) error {
	fmt.Println("Entering a temporary shell...")
	args := append([]string{"-p"}, packages...)
	return nix.RunInteractiveCommand("nix-shell", args...)
}

// Remove removes packages from the user profile.
func Remove(packages []string) error {
	if nix.GetNixMode() == nix.NixOS {
		return fmt.Errorf("this command is not supported on NixOS")
	}
	fmt.Println("Removing packages from your user profile...")
	args := append([]string{"profile", "remove"}, packages...)
	_, err := nix.RunCommand("nix", args...)
	return err
}

func commitChanges(message string) error {
	path := config.GetInstallPath()
	if err := GitAdd(path); err != nil {
		return fmt.Errorf("could not add changes: %w", err)
	}
	if err := GitCommit(path, message); err != nil {
		return fmt.Errorf("could not commit changes: %w", err)
	}
	return nil
}
