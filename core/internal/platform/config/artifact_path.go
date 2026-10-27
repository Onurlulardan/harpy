package config

import (
	"os"
	"path/filepath"
	"runtime"
)

type ArtifactAccess string

const (
	ArtifactReadOnly        ArtifactAccess = "read_only"
	ArtifactReadOnceCleanup ArtifactAccess = "read_once_cleanup"
	ArtifactCreateOnly      ArtifactAccess = "create_only"
	ArtifactAtomicWrite     ArtifactAccess = "atomic_write"
	ArtifactReadWrite       ArtifactAccess = "read_write"
)

type ArtifactPath struct {
	path   AbsolutePath
	access ArtifactAccess
}

func (path ArtifactPath) Path() AbsolutePath     { return path.path }
func (path ArtifactPath) Access() ArtifactAccess { return path.access }

func loadReadArtifact(
	values environmentValues,
	key string,
	sensitive bool,
) (ArtifactPath, error) {
	return loadArtifactPath(values, key, ArtifactReadOnly, sensitive, true)
}

func loadOptionalReadArtifact(
	values environmentValues,
	key string,
) (ArtifactPath, bool, error) {
	if _, found := values.lookup(key); !found {
		return ArtifactPath{}, false, nil
	}
	path, err := loadReadArtifact(values, key, false)
	return path, err == nil, err
}

func loadArtifactPath(
	values environmentValues,
	key string,
	access ArtifactAccess,
	sensitive bool,
	mustExist bool,
) (ArtifactPath, error) {
	path, err := loadAbsolutePath(values, key)
	if err != nil {
		return ArtifactPath{}, err
	}
	resolved, info, err := inspectArtifactPath(path.Value(), mustExist)
	if err != nil {
		return ArtifactPath{}, configError(key, err.Error())
	}
	if info != nil {
		if access == ArtifactCreateOnly {
			if !info.IsDir() {
				return ArtifactPath{}, configError(key, "must resolve to a directory")
			}
		} else if !info.Mode().IsRegular() {
			return ArtifactPath{}, configError(key, "must resolve to a regular file")
		}
		if err := validateArtifactPermissions(info, sensitive); err != nil {
			return ArtifactPath{}, configError(key, err.Error())
		}
	}
	clean, valid := newAbsolutePath(resolved)
	if !valid {
		return ArtifactPath{}, configError(key, "must resolve to an absolute clean path")
	}
	return ArtifactPath{path: clean, access: access}, nil
}

func inspectArtifactPath(path string, mustExist bool) (string, os.FileInfo, error) {
	_, err := os.Lstat(path)
	if err == nil {
		resolved, resolveErr := filepath.EvalSymlinks(path)
		if resolveErr != nil {
			return "", nil, artifactPathError("cannot be resolved")
		}
		resolvedInfo, statErr := os.Stat(resolved)
		if statErr != nil {
			return "", nil, artifactPathError("cannot be inspected")
		}
		return filepath.Clean(resolved), resolvedInfo, nil
	}
	if !os.IsNotExist(err) || mustExist {
		return "", nil, artifactPathError("does not exist")
	}
	parent, resolveErr := filepath.EvalSymlinks(filepath.Dir(path))
	if resolveErr != nil {
		return "", nil, artifactPathError("parent directory cannot be resolved")
	}
	parentInfo, statErr := os.Stat(parent)
	if statErr != nil || !parentInfo.IsDir() {
		return "", nil, artifactPathError("parent must be a directory")
	}
	if err := validateArtifactPermissions(parentInfo, false); err != nil {
		return "", nil, err
	}
	return filepath.Join(parent, filepath.Base(path)), nil, nil
}

type artifactPathError string

func (err artifactPathError) Error() string { return string(err) }

func validateArtifactPermissions(info os.FileInfo, sensitive bool) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	if sensitive && info.Mode().Perm()&0o077 != 0 {
		return artifactPathError("permissions must be owner-only")
	}
	if !sensitive && info.Mode().Perm()&0o022 != 0 {
		return artifactPathError("permissions are too broad")
	}
	return nil
}

func validateDistinctArtifactPaths(key string, paths ...ArtifactPath) error {
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		value := path.Path().Value()
		if _, duplicate := seen[value]; duplicate {
			return configError(key, "artifact paths must be distinct")
		}
		seen[value] = struct{}{}
	}
	return nil
}
