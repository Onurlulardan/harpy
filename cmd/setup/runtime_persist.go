package main

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
)

// setupRotateSecrets is set by run for one setup invocation.
// Direct runtime and release calls keep installed files.
// setupRefreshRelease rebuilds the release bundle without rotating runtime keys.
var (
	setupRotateSecrets  bool
	setupRefreshRelease bool
)

func writeDurableFile(dir, name string, data []byte) (string, error) {
	return writeMaterialFile(dir, name, data, true)
}

func writeMaterialFile(dir, name string, data []byte, keep bool) (string, error) {
	path := filepath.Join(dir, name)
	if keep {
		exists, err := installedFile(path)
		if err != nil {
			return "", err
		}
		if exists {
			return path, nil
		}
	}
	if err := writePrivate(path, data); err != nil {
		return "", err
	}
	return path, nil
}

func installedFile(path string) (bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Size() == 0 {
		return false, nil
	}
	return true, nil
}

func writeRuntimeGroup(dir string, bodies map[string][]byte) (map[string]string, error) {
	names := make([]string, 0, len(bodies))
	for name := range bodies {
		names = append(names, name)
	}
	sort.Strings(names)
	keep := !setupRotateSecrets
	if keep {
		complete, err := groupInstalled(dir, names)
		if err != nil {
			return nil, err
		}
		keep = complete
	}
	paths := make(map[string]string, len(names))
	for _, name := range names {
		path, err := writeMaterialFile(dir, name, bodies[name], keep)
		if err != nil {
			return nil, err
		}
		paths[name] = path
	}
	return paths, nil
}

func groupInstalled(dir string, names []string) (bool, error) {
	for _, name := range names {
		exists, err := installedFile(filepath.Join(dir, name))
		if err != nil || !exists {
			return false, err
		}
	}
	return true, nil
}
