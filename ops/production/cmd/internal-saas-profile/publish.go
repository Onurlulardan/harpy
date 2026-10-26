package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

type publishResult struct {
	Output    string
	Generated bool
}

func ensureProfile(options profileOptions) (publishResult, error) {
	validated, err := options.validate()
	if err != nil {
		return publishResult{}, err
	}
	files, err := renderProfiles(validated)
	if err != nil {
		return publishResult{}, err
	}
	manifest, err := buildManifest(validated, files)
	if err != nil {
		return publishResult{}, fmt.Errorf("build manifest: %w", err)
	}
	files[manifestName] = manifest
	match, exists, err := profileSetMatches(validated.output, files)
	if err != nil {
		return publishResult{}, err
	}
	if exists {
		if !match {
			return publishResult{}, fmt.Errorf(
				"internal SaaS profile exists with different or invalid content")
		}
		return publishResult{Output: validated.output}, nil
	}
	if err := publishProfileSet(validated.output, files); err != nil {
		return publishResult{}, err
	}
	return publishResult{Output: validated.output, Generated: true}, nil
}

func profileSetMatches(output string, files map[string][]byte) (bool, bool, error) {
	info, err := os.Lstat(output)
	if os.IsNotExist(err) {
		return false, false, nil
	}
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false, true, fmt.Errorf("internal SaaS profile output is unsafe")
	}
	entries, err := os.ReadDir(output)
	if err != nil || len(entries) != len(files) {
		return false, true, err
	}
	for _, entry := range entries {
		expected, found := files[entry.Name()]
		if !found || entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return false, true, nil
		}
		actual, readErr := os.ReadFile(filepath.Join(output, entry.Name()))
		if readErr != nil {
			return false, true, readErr
		}
		if !bytes.Equal(actual, expected) {
			return false, true, nil
		}
	}
	return true, true, nil
}

func publishProfileSet(output string, files map[string][]byte) (err error) {
	parent := filepath.Dir(output)
	if err = os.MkdirAll(parent, 0o700); err != nil {
		return fmt.Errorf("create profile parent: %w", err)
	}
	staged, err := os.MkdirTemp(parent, ".internal-saas-profile-")
	if err != nil {
		return fmt.Errorf("stage profile: %w", err)
	}
	defer func() { _ = os.RemoveAll(staged) }()
	if err = os.Chmod(staged, 0o700); err != nil {
		return fmt.Errorf("restrict staged profile: %w", err)
	}
	for name, content := range files {
		if err = os.WriteFile(filepath.Join(staged, name), content, 0o600); err != nil {
			return fmt.Errorf("write profile file %s: %w", name, err)
		}
	}
	if err = os.Rename(staged, output); err != nil {
		return fmt.Errorf("publish profile: %w", err)
	}
	return nil
}
