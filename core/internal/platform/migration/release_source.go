package migration

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

const releaseArchiveName = "migration-release.zip"

type releaseReader interface {
	names() ([]string, error)
	read(name string) ([]byte, error)
}

func openDefaultReleaseArchive() (releaseReader, error) {
	if zipPath := zipBesideExecutable(); zipPath != "" {
		return openZipRelease(zipPath)
	}
	directory, err := findReleaseDirectory()
	if err != nil {
		return nil, err
	}
	return directoryRelease{root: directory}, nil
}

func zipBesideExecutable() string {
	executable, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		executable = resolved
	}
	candidate := filepath.Join(filepath.Dir(executable), releaseArchiveName)
	info, err := os.Stat(candidate)
	if err != nil || info.IsDir() {
		return ""
	}
	return candidate
}

func findReleaseDirectory() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	if releaseDirOK(filepath.Join(cwd, "release")) {
		return filepath.Join(cwd, "release"), nil
	}
	dir := cwd
	for {
		candidate := filepath.Join(dir, "core", "internal", "platform", "migration", "release")
		if releaseDirOK(candidate) {
			return candidate, nil
		}
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("migration release archive is unavailable")
}

func releaseDirOK(directory string) bool {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), releaseManifestSuffix) {
			return true
		}
	}
	return false
}

type directoryRelease struct {
	root string
}

func (archive directoryRelease) names() ([]string, error) {
	entries, err := os.ReadDir(archive.root)
	if err != nil {
		return nil, fmt.Errorf("read migration release archive: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			return nil, fmt.Errorf("migration release archive contains a directory")
		}
		names = append(names, entry.Name())
	}
	return names, nil
}

func (archive directoryRelease) read(name string) ([]byte, error) {
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return nil, fmt.Errorf("migration manifest name is invalid")
	}
	body, err := os.ReadFile(filepath.Join(archive.root, name))
	if err != nil {
		return nil, fmt.Errorf("read migration manifest: %w", err)
	}
	return limitManifest(body)
}

type zipRelease struct {
	reader *zip.ReadCloser
}

func openZipRelease(zipPath string) (*zipRelease, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, fmt.Errorf("open migration release archive: %w", err)
	}
	return &zipRelease{reader: reader}, nil
}

func (archive *zipRelease) names() ([]string, error) {
	names := make([]string, 0, len(archive.reader.File))
	for _, file := range archive.reader.File {
		if file.FileInfo().IsDir() {
			continue
		}
		names = append(names, path.Base(file.Name))
	}
	return names, nil
}

func (archive *zipRelease) read(name string) ([]byte, error) {
	for _, file := range archive.reader.File {
		if path.Base(file.Name) != name {
			continue
		}
		opened, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("read migration manifest: %w", err)
		}
		body, err := io.ReadAll(io.LimitReader(opened, int64(postgresv1.MaximumMigrationManifestBytes)+1))
		closeErr := opened.Close()
		if err != nil {
			return nil, fmt.Errorf("read migration manifest: %w", err)
		}
		if closeErr != nil {
			return nil, closeErr
		}
		return limitManifest(body)
	}
	return nil, fmt.Errorf("migration manifest is not in the release archive")
}

func limitManifest(body []byte) ([]byte, error) {
	if len(body) == 0 || len(body) > postgresv1.MaximumMigrationManifestBytes {
		return nil, fmt.Errorf("migration manifest size is invalid")
	}
	return body, nil
}
