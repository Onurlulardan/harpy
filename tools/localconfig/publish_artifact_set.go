package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func publishArtifactSet(
	output string,
	set *artifactSet,
	manifest setManifest,
) error {
	parent := filepath.Dir(output)
	staged, err := os.MkdirTemp(parent, ".localconfig-staged-")
	if err != nil {
		return fmt.Errorf("create staged artifact set: %w", err)
	}
	published := false
	defer func() {
		if !published {
			_ = os.RemoveAll(staged)
		}
	}()
	for _, name := range set.expectedArtifactNames() {
		if err := writeArtifactFile(filepath.Join(staged, name), set.files[name]); err != nil {
			return err
		}
	}
	manifestContent, err := marshalJSON(manifest)
	if err != nil {
		return err
	}
	if err := writeArtifactFile(
		filepath.Join(staged, manifestFileName), manifestContent,
	); err != nil {
		return err
	}
	generatedAt, err := parseCanonicalTimestamp(manifest.GeneratedAt)
	if err != nil {
		return err
	}
	_, current, inspectErr := inspectPublishedArtifactSet(staged, generatedAt, set)
	if inspectErr != nil {
		return fmt.Errorf("inspect staged artifact set: %w", inspectErr)
	}
	if !current {
		return fmt.Errorf("staged artifact set validation failed")
	}
	if _, err := os.Lstat(output); err == nil {
		return fmt.Errorf("publication target already exists")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect publication target: %w", err)
	}
	if err := os.Rename(staged, output); err != nil {
		return fmt.Errorf("publish artifact set: %w", err)
	}
	published = true
	return nil
}

func inspectPublishedArtifactSet(
	path string,
	generatedAt time.Time,
	set *artifactSet,
) (setManifest, bool, error) {
	if set.options.profile == profileForgeGuard {
		return inspectForgeGuardArtifactSet(
			path, generatedAt, set.options, false,
		)
	}
	return inspectArtifactSet(path, generatedAt, false)
}

func writeArtifactFile(path string, content []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create artifact %s: %w", filepath.Base(path), err)
	}
	if _, err := file.Write(content); err != nil {
		_ = file.Close()
		return fmt.Errorf("write artifact %s: %w", filepath.Base(path), err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync artifact %s: %w", filepath.Base(path), err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close artifact %s: %w", filepath.Base(path), err)
	}
	return nil
}
