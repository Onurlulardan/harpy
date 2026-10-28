package migrationbundle

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const releaseManifestSuffix = ".migration-manifest.json"

func PackRelease(root, output string) error {
	if err := ReleaseNamesMatchIndex(root); err != nil {
		return err
	}
	directory := releaseDirectory(root)
	entries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	writer := zip.NewWriter(file)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), releaseManifestSuffix) {
			continue
		}
		body, readErr := os.ReadFile(filepath.Join(directory, entry.Name()))
		if readErr != nil {
			_ = writer.Close()
			_ = file.Close()
			return readErr
		}
		member, createErr := writer.Create(entry.Name())
		if createErr != nil {
			_ = writer.Close()
			_ = file.Close()
			return createErr
		}
		if _, writeErr := member.Write(body); writeErr != nil {
			_ = writer.Close()
			_ = file.Close()
			return writeErr
		}
	}
	if err = writer.Close(); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func ReleaseNamesMatchIndex(root string) error {
	digests, err := indexedDigests(root)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(releaseDirectory(root))
	if err != nil {
		return err
	}
	found := map[string]struct{}{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), releaseManifestSuffix) {
			continue
		}
		found[strings.TrimSuffix(entry.Name(), releaseManifestSuffix)] = struct{}{}
	}
	if len(found) != len(digests) {
		return fmt.Errorf("release archive names do not match the release index")
	}
	for _, digest := range digests {
		if _, exists := found[digest]; !exists {
			return fmt.Errorf("release archive names do not match the release index")
		}
	}
	return nil
}

func indexedDigests(root string) ([]string, error) {
	body, err := os.ReadFile(filepath.Join(
		root, "core", "internal", "platform", "migration", "release-index.json",
	))
	if err != nil {
		return nil, err
	}
	var index struct {
		Digests []string `json:"digests"`
	}
	if err = json.Unmarshal(body, &index); err != nil {
		return nil, err
	}
	if len(index.Digests) == 0 {
		return nil, fmt.Errorf("release index has no digests")
	}
	return index.Digests, nil
}

func releaseDirectory(root string) string {
	return filepath.Join(root, "core", "internal", "platform", "migration", "release")
}
