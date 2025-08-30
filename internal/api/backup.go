package api

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// GitBackup creates a compressed tarball of the repository at repoPath.
// The backup is stored in a .backups directory within the repository.
// The .git directory and the .backups directory itself are excluded from the backup.
func GitBackup(repoPath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	backupsDir := filepath.Join(home, ".local", "share", "pilo", "backups")
	if err := os.MkdirAll(backupsDir, 0755); err != nil {
		return fmt.Errorf("failed to create backups directory: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	backupFileName := fmt.Sprintf("backup-%s.tar.gz", timestamp)
	backupFilePath := filepath.Join(backupsDir, backupFileName)

	file, err := os.Create(backupFilePath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Exclude .git and .backups directories
		if info.IsDir() && (info.Name() == ".git" || info.Name() == ".backups") {
			return filepath.SkipDir
		}

		// Create a relative path for the tar header
		relPath, err := filepath.Rel(repoPath, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, relPath)
		if err != nil {
			return err
		}
		header.Name = strings.ReplaceAll(relPath, string(filepath.Separator), "/") // Use forward slashes in tar header

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(tarWriter, f); err != nil {
			return err
		}

		return nil
	})
}

// RestoreMostRecentBackup finds the most recent backup and restores it to the given path.
func RestoreMostRecentBackup(repoPath string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	backupsDir := filepath.Join(home, ".local", "share", "pilo", "backups")

	files, err := os.ReadDir(backupsDir)
	if err != nil {
		return fmt.Errorf("failed to read backups directory: %w", err)
	}

	var mostRecentFile string
	var mostRecentTime time.Time

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".tar.gz") {
			info, err := file.Info()
			if err != nil {
				continue // Skip files we can't get info for
			}
			if mostRecentFile == "" || info.ModTime().After(mostRecentTime) {
				mostRecentTime = info.ModTime()
				mostRecentFile = file.Name()
			}
		}
	}

	if mostRecentFile == "" {
		return errors.New("no backups found")
	}

	backupFilePath := filepath.Join(backupsDir, mostRecentFile)

	// Clear the destination directory before restoring
	if err := os.RemoveAll(repoPath); err != nil {
		return fmt.Errorf("failed to clear destination directory before restore: %w", err)
	}
	if err := os.MkdirAll(repoPath, 0755); err != nil {
		return fmt.Errorf("failed to recreate destination directory: %w", err)
	}

	file, err := os.Open(backupFilePath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of tar archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		target := filepath.Join(repoPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory from backup: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for file: %w", err)
			}
			outFile, err := os.Create(target)
			if err != nil {
				return fmt.Errorf("failed to create file from backup: %w", err)
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file content from backup: %w", err)
			}
			outFile.Close()
		}
	}

	return nil
}
