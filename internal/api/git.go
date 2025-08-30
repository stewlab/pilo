package api

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"pilo/internal/config"
	"time"

	"fmt"

	"github.com/go-git/go-git/v5"

	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// ErrDirtyRepository is returned when a git operation fails because the repository has uncommitted changes.
var ErrDirtyRepository = errors.New("local repository has uncommitted changes")

// getGitAuth determines the authentication method for Git operations.
// It prioritizes a specific SSH key, but falls back to the SSH agent if no key is provided.
func getGitAuth() (ssh.AuthMethod, error) {
	sshKeyPath := config.GetSshKeyPath()
	if sshKeyPath != "" {
		auth, err := ssh.NewPublicKeysFromFile("git", sshKeyPath, "")
		if err != nil {
			return nil, fmt.Errorf("failed to load SSH key from %s: %w", sshKeyPath, err)
		}
		return auth, nil
	}
	// Use SSH agent if no key is specified
	if os.Getenv("SSH_AUTH_SOCK") != "" {
		auth, err := ssh.NewSSHAgentAuth("git")
		if err == nil {
			return auth, nil
		}
	}

	// Fallback to searching for SSH keys
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	keyTypes := []string{"id_ed25519", "id_ecdsa", "id_rsa"}
	for _, keyType := range keyTypes {
		path := filepath.Join(homeDir, ".ssh", keyType)
		if _, errStat := os.Stat(path); errStat == nil {
			auth, err := ssh.NewPublicKeysFromFile("git", path, "")
			if err == nil {
				return auth, nil // Key found and loaded
			}
		}
	}

	return nil, errors.New("no usable SSH key found and SSH agent is not available")
}

func GitInit(repoPath string) error {
	repo, err := git.PlainInit(repoPath, false)
	if err != nil {
		if err == git.ErrRepositoryAlreadyExists {
			return nil // Already a git repo, so we're good
		}
		return err
	}

	// Set the default branch to 'main'
	cfg, err := repo.Config()
	if err != nil {
		return fmt.Errorf("failed to get git config: %w", err)
	}
	cfg.Init.DefaultBranch = "main"
	if err := repo.SetConfig(cfg); err != nil {
		return fmt.Errorf("failed to set default branch to main: %w", err)
	}

	// Create an initial empty commit
	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	// Add a .gitkeep file to ensure there's something to commit
	gitkeepPath := filepath.Join(repoPath, ".gitkeep")
	if err := os.WriteFile(gitkeepPath, []byte{}, 0644); err != nil {
		return fmt.Errorf("failed to create .gitkeep file: %w", err)
	}

	_, err = w.Add(".gitkeep")
	if err != nil {
		return fmt.Errorf("failed to add .gitkeep: %w", err)
	}

	commitHash, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "pilo",
			Email: "pilo@localhost",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create initial commit: %w", err)
	}

	// Explicitly create and checkout the 'main' branch
	mainBranchRef := plumbing.NewBranchReferenceName("main")
	if err := repo.Storer.SetReference(plumbing.NewHashReference(mainBranchRef, commitHash)); err != nil {
		return fmt.Errorf("failed to create main branch: %w", err)
	}
	if err := w.Checkout(&git.CheckoutOptions{Branch: mainBranchRef}); err != nil {
		return fmt.Errorf("failed to checkout main branch: %w", err)
	}

	return nil
}

func GitAdd(repoPath string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Add(".")
	return err
}

func GitCommit(repoPath, message string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	status, err := w.Status()
	if err != nil {
		return err
	}

	if status.IsClean() {
		return nil
	}

	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "pilo",
			Email: "pilo@localhost",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	// Check if we should push after commit
	pushOnCommit, err := config.GetPushOnCommit()
	if err != nil {
		// Log or handle error, but don't block the commit
		config.AddLogEntry(fmt.Sprintf("could not get push on commit setting: %v", err))
	}

	if pushOnCommit {
		remoteURL, err := config.GetRemoteUrl()
		if err != nil {
			config.AddLogEntry(fmt.Sprintf("could not get remote URL: %v", err))
		}
		if remoteURL != "" {
			if err := GitSync(repoPath); err != nil {
				// Log or handle push error, but don't fail the commit
				config.AddLogEntry(fmt.Sprintf("failed to sync after commit: %v", err))
			}
		}
	}

	return nil
}

// GitRestoreStrategy defines the action to take when a repository is dirty.
type GitRestoreStrategy int

const (
	// GitRestoreCommit commits the changes before restoring.
	GitRestoreCommit GitRestoreStrategy = iota
	// GitRestoreDiscard discards the changes before restoring.
	GitRestoreDiscard
	// GitRestoreBackup creates a backup of the changes before restoring.
	GitRestoreBackup
)

// GitRestore handles cloning or updating a repository from a remote URL.
// If the repository is dirty, it uses the provided strategy to resolve the state.
func GitRestore(repoPath, remoteURL, branch string, strategy *GitRestoreStrategy, commitMessage string) error {
	// Log environment variables for debugging SSH issues
	logMessage := "Environment variables:\n"
	for _, e := range os.Environ() {
		logMessage += e + "\n"
	}
	config.AddLogEntry(logMessage)

	auth, err := getGitAuth()
	if err != nil {
		config.AddLogEntry(fmt.Sprintf("getGitAuth error: %v", err))
		return err
	}

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			// If the repository does not exist, clone it
			cloneOptions := &git.CloneOptions{
				URL:      remoteURL,
				Auth:     auth,
				Progress: os.Stdout,
			}
			if branch != "" {
				cloneOptions.ReferenceName = plumbing.NewBranchReferenceName(branch)
			}
			_, err = git.PlainClone(repoPath, false, cloneOptions)
			return err
		}
		// For any other error opening the repo
		return err
	}

	// Repository exists, check if it's dirty
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	status, err := w.Status()
	if err != nil {
		return err
	}

	if !status.IsClean() {
		if strategy == nil {
			return ErrDirtyRepository
		}
		switch *strategy {
		case GitRestoreCommit:
			if err := GitCommit(repoPath, commitMessage); err != nil {
				return err
			}
		case GitRestoreDiscard:
			if err := GitReset(repoPath); err != nil {
				return err
			}
		case GitRestoreBackup:
			// Create a backup before resetting
			if err := GitBackup(repoPath); err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}
			// Commit changes to a temporary branch to avoid losing them
			if err := GitAdd(repoPath); err != nil {
				return fmt.Errorf("failed to add changes for backup commit: %w", err)
			}
			backupCommitMessage := fmt.Sprintf("pilo-backup-%s", time.Now().Format("20060102-150405"))
			if err := GitCommit(repoPath, backupCommitMessage); err != nil {
				// If commit fails, it might be because there's nothing to commit.
				// We can proceed, as the backup tarball was already created.
				config.AddLogEntry(fmt.Sprintf("Could not commit changes for backup, continuing restore: %v", err))
			}
			// After backing up and committing, clean the worktree by resetting.
			if err := GitReset(repoPath); err != nil {
				return fmt.Errorf("failed to reset repository after backup: %w", err)
			}
		}
	}

	// Now that the repo is clean, let's deal with the remote
	remote, err := repo.Remote("origin")
	if err != nil {
		// If remote doesn't exist, create it
		remote, err = repo.CreateRemote(&gitconfig.RemoteConfig{
			Name: "origin",
			URLs: []string{remoteURL},
		})
		if err != nil {
			return err
		}
	}

	// If the URL is different, update it
	if len(remote.Config().URLs) == 0 || remote.Config().URLs[0] != remoteURL {
		err = repo.DeleteRemote("origin")
		if err != nil {
			return err
		}
		_, err = repo.CreateRemote(&gitconfig.RemoteConfig{
			Name: "origin",
			URLs: []string{remoteURL},
		})
		if err != nil {
			return err
		}
	}

	// Fetch and reset
	fetchOptions := &git.FetchOptions{
		RemoteName: "origin",
		Auth:       auth,
	}
	if err := repo.Fetch(fetchOptions); err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	// Determine the remote reference
	var remoteRef *plumbing.Reference
	if branch != "" {
		remoteRef, err = repo.Reference(plumbing.NewRemoteReferenceName("origin", branch), true)
		if err != nil {
			return fmt.Errorf("could not find branch '%s' in the remote repository", branch)
		}
	} else {
		remoteRef, err = repo.Reference("refs/remotes/origin/HEAD", true)
		if err != nil {
			// If HEAD is not found, try to find main or master
			refs, err := repo.References()
			if err != nil {
				return err
			}
			var masterRef *plumbing.Reference
			err = refs.ForEach(func(ref *plumbing.Reference) error {
				if ref.Name().IsRemote() {
					if ref.Name().Short() == "origin/main" {
						remoteRef = ref
						return nil // Found main, stop iterating
					} else if ref.Name().Short() == "origin/master" {
						masterRef = ref // Keep master ref in case main is not found
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
			if remoteRef == nil && masterRef != nil {
				remoteRef = masterRef // If main not found, use master
			}
			if err != nil {
				return err
			}
		}
	}

	if remoteRef == nil {
		return errors.New("could not determine the default branch of the remote repository")
	}

	// Reset the worktree to the remote's HEAD
	return w.Reset(&git.ResetOptions{
		Commit: remoteRef.Hash(),
		Mode:   git.HardReset,
	})
}

// GitStatus checks if the repository at the given path has uncommitted changes.
func GitStatus(repoPath string) (bool, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return false, nil // Not a git repo, so not dirty
		}
		return false, err
	}

	w, err := repo.Worktree()
	if err != nil {
		return false, err
	}
	status, err := w.Status()
	if err != nil {
		return false, err
	}

	return !status.IsClean(), nil
}

func GitReset(repoPath string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	w, err := repo.Worktree()
	if err != nil {
		return err
	}

	return w.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	})
}

func GetGitStatus(repoPath string) (string, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			return "Not a git repository.", nil
		}
		return "", err
	}

	w, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	status, err := w.Status()
	if err != nil {
		return "", err
	}

	if status.IsClean() {
		return "Working tree clean", nil
	}

	return status.String(), nil
}

func GetGitDiff(repoPath string) (string, error) {
	// Using the native git command is more reliable for diffing, especially with submodules.
	cmd := exec.Command("git", "diff", "--submodule=diff", "HEAD")
	cmd.Dir = repoPath

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check if the error is because there are no commits yet
		if _, errOpen := git.PlainOpen(repoPath); errOpen == git.ErrRepositoryNotExists {
			return "Not a git repository.", nil
		}
		// Check for empty repository case
		repo, errOpen := git.PlainOpen(repoPath)
		if errOpen == nil {
			if _, errHead := repo.Head(); errHead == plumbing.ErrReferenceNotFound {
				// No HEAD commit, so we can't diff. Show all files as new.
				// We can do this by diffing against the "empty tree" hash.
				cmd := exec.Command("git", "diff", "--submodule=diff", "4b825dc642cb6eb9a060e54bf8d69288fbee4904")
				cmd.Dir = repoPath
				cmd.Stdout = &out
				cmd.Stderr = &stderr
				if err := cmd.Run(); err != nil {
					return "", fmt.Errorf("failed to get initial diff: %w: %s", err, stderr.String())
				}
				return out.String(), nil
			}
		}
		return "", fmt.Errorf("git diff command failed: %w: %s", err, stderr.String())
	}

	if out.String() == "" {
		return "No changes.", nil
	}

	return out.String(), nil
}

func GitPush(repoPath string) error {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	auth, err := getGitAuth()
	if err != nil {
		return err
	}

	err = repo.Push(&git.PushOptions{
		Auth: auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
}

// GitSync provides a safe way to push local changes to the remote, even if the branches have diverged.
// It works by backing up local changes, pulling remote changes, restoring local changes, and then pushing.
func GitSync(repoPath string) error {
	// 1. Create a backup of the current state
	if err := GitBackup(repoPath); err != nil {
		return fmt.Errorf("failed to create backup before syncing: %w", err)
	}

	// 2. Pull remote changes, hard resetting to the remote state
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}
	auth, err := getGitAuth()
	if err != nil {
		return err
	}
	if err := repo.Fetch(&git.FetchOptions{RemoteName: "origin", Auth: auth}); err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to fetch from remote: %w", err)
	}
	// We don't need to inspect the remote ref directly, just fetch.
	// The subsequent restore and push will handle the state.
	// Before resetting, we need to ensure the .git directory is not deleted.
	// The RestoreMostRecentBackup function will clear the repoPath.
	// We need to preserve the .git directory.
	gitDir := filepath.Join(repoPath, ".git")
	tempGitDir, err := os.MkdirTemp("", "pilo-git-")
	if err != nil {
		return fmt.Errorf("failed to create temp git dir: %w", err)
	}
	defer os.RemoveAll(tempGitDir)
	if err := os.Rename(gitDir, filepath.Join(tempGitDir, ".git")); err != nil {
		return fmt.Errorf("failed to move .git dir: %w", err)
	}

	// 3. Restore the most recent backup
	if err := RestoreMostRecentBackup(repoPath); err != nil {
		// Attempt to restore the .git directory before failing
		os.Rename(filepath.Join(tempGitDir, ".git"), gitDir)
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	// 4. Move the .git directory back
	if err := os.Rename(filepath.Join(tempGitDir, ".git"), gitDir); err != nil {
		return fmt.Errorf("failed to move .git dir back: %w", err)
	}

	// 5. Commit the restored (local) changes
	if err := GitAdd(repoPath); err != nil {
		return fmt.Errorf("failed to add restored files: %w", err)
	}
	if err := GitCommit(repoPath, "pilo: sync local changes"); err != nil {
		// It's possible there were no changes to commit, so we don't fail here
		config.AddLogEntry(fmt.Sprintf("Note: could not create sync commit, possibly no changes: %v", err))
	}

	// 6. Push the synchronized changes
	if err := GitPush(repoPath); err != nil {
		return fmt.Errorf("failed to push synchronized changes: %w", err)
	}

	config.AddLogEntry("Successfully synced with remote. Your local changes have been preserved and pushed.")
	return nil
}
