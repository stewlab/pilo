package api

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"pilo/internal/config"
	"slices"
	"strings"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

var FlakeFS embed.FS

func SetFlakeFS(flakeFS embed.FS) {
	FlakeFS = flakeFS
}

func must(s string, e error) string {
	if e != nil {
		panic(e)
	}
	return s
}

func Inflate(targetPath string, remoteURL string, cleanTargetPath bool) error {
	// For internal installs, commit any pre-install changes and clean the directory
	if remoteURL == "" {
		if err := GitAdd(targetPath); err != nil {
			return err
		}
		if err := GitCommit(targetPath, "pilo: pre-install changes"); err != nil {
			return err
		}
	}

	fmt.Println("Before cleanDir")
	if cleanTargetPath || remoteURL == "" {
		ignoreList := []string{".backups", ".git", ".gitignore"}
		if err := cleanDir(targetPath, ignoreList); err != nil {
			return err
		}
	}
	fmt.Println("After cleanDir")

	// Check if we should install from remote
	if remoteURL != "" {
		var publicKeys *ssh.PublicKeys
		var err error
		sshKeyPath := config.GetSshKeyPath()
		if sshKeyPath != "" {
			publicKeys, err = ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
			if err != nil {
				return fmt.Errorf("failed to load SSH key from %s: %w", sshKeyPath, err)
			}
		} else {
			// Find an available SSH key
			homeDir := must(os.UserHomeDir())
			keyTypes := []string{"id_ed25519", "id_ecdsa", "id_rsa"}
			for _, keyType := range keyTypes {
				path := filepath.Join(homeDir, ".ssh", keyType)
				if _, errStat := os.Stat(path); errStat == nil {
					publicKeys, err = ssh.NewPublicKeysFromFile("git", path, "")
					if err == nil {
						break // Key found and loaded
					}
				}
			}
			if publicKeys == nil {
				return fmt.Errorf("no usable SSH key found in ~/.ssh")
			}
		}

		repo, err := git.PlainOpen(targetPath)
		// If repo exists, fetch and reset
		if err == nil {
			// Commit any pre-install changes before fetching/resetting
			if err := GitAdd(targetPath); err != nil {
				return err
			}
			if err := GitCommit(targetPath, "pilo: pre-reinstall changes"); err != nil {
				return err
			}

			// Add a new remote
			_, err := repo.CreateRemote(&gitconfig.RemoteConfig{
				Name: "new-origin",
				URLs: []string{remoteURL},
			})
			// if remote already exists, just fetch
			if err != nil && err != git.ErrRemoteExists {
				return err
			}

			// Fetch from the new remote
			if err := repo.Fetch(&git.FetchOptions{
				RemoteName: "new-origin",
				Progress:   os.Stdout,
				Auth:       publicKeys,
			}); err != nil && err.Error() != "already up-to-date" {
				return err
			}

			// Get the worktree
			w, err := repo.Worktree()
			if err != nil {
				return err
			}

			// Get the hash of the remote's main branch
			remoteRef, err := repo.Reference("refs/remotes/new-origin/main", true)
			if err != nil {
				return err
			}

			// Reset to the fetched branch
			if err := w.Reset(&git.ResetOptions{
				Mode:   git.HardReset,
				Commit: remoteRef.Hash(),
			}); err != nil {
				return err
			}
		} else if err == git.ErrRepositoryNotExists {
			// If repo doesn't exist, clone it
			if _, err := git.PlainClone(targetPath, false, &git.CloneOptions{
				URL:      remoteURL,
				Progress: os.Stdout,
				Auth:     publicKeys,
			}); err != nil {
				return err
			}
			// Create a default base-config.json after cloning
			configPath := filepath.Join(targetPath, "flake", "base-config.json")
			if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
				return err
			}
			if err := os.WriteFile(configPath, []byte("{\n  \"commit_triggers\": [],\n  \"remote_url\": \"\",\n  \"push_on_commit\": false\n}\n"), 0644); err != nil {
				return err
			}
		} else {
			// Other errors
			return err
		}
	} else {
		// Inflate the flake from embedded FS
		flakeRoot := "flake"
		if err := fs.WalkDir(FlakeFS, flakeRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			// Create a relative path from the flakeRoot to preserve the directory structure.
			relativePath, err := filepath.Rel(flakeRoot, path)
			if err != nil {
				return err
			}
			target := filepath.Join(targetPath, "flake", relativePath)

			if d.IsDir() {
				return os.MkdirAll(target, 0755)
			}

			data, err := FlakeFS.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(target, data, 0644)
		}); err != nil {
			return err
		}
	}

	// Create or update the .gitignore file after inflating the flake.
	if err := CreateGitignore(targetPath); err != nil {
		return err
	}

	// Commit post-install changes
	if err := GitAdd(targetPath); err != nil {
		return err
	}
	return GitCommit(targetPath, "pilo: post-install changes")
}

// CreateGitignore creates or updates a .gitignore file at the specified path.
func CreateGitignore(path string) error {
	gitignorePath := filepath.Join(path, ".gitignore")
	backupsEntry := "/.backups/"

	// Check if the file exists
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		// File does not exist, create it with the backups entry
		return os.WriteFile(gitignorePath, []byte(backupsEntry+"\n"), 0644)
	} else if err != nil {
		// Another error occurred
		return fmt.Errorf("error checking .gitignore: %w", err)
	}

	// File exists, read it and check if the backups entry is already there
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		return fmt.Errorf("error reading .gitignore: %w", err)
	}

	if !strings.Contains(string(content), backupsEntry) {
		// Entry does not exist, append it
		f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error opening .gitignore for append: %w", err)
		}
		defer f.Close()

		if _, err := f.WriteString("\n" + backupsEntry + "\n"); err != nil {
			return fmt.Errorf("error appending to .gitignore: %w", err)
		}
	}

	return nil
}

func cleanDir(path string, ignoreList []string) error {
	gitPath := filepath.Join(path, ".git")
	if _, err := os.Stat(gitPath); os.IsNotExist(err) {
		// Not a git repository, so don't clean
		return nil
	}

	d, err := os.Open(path)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		if !slices.Contains(ignoreList, name) {
			err = os.RemoveAll(filepath.Join(path, name))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
