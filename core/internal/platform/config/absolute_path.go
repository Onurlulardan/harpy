package config

import (
	"path/filepath"
	"strings"
)

type AbsolutePath struct {
	value string
}

func loadAbsolutePath(values environmentValues, key string) (AbsolutePath, error) {
	value, found := values.lookup(key)
	if !found || value == "" {
		return AbsolutePath{}, configError(key, "is required")
	}
	path, valid := newAbsolutePath(value)
	if !valid {
		return AbsolutePath{}, configError(key, "must be absolute and clean")
	}
	return path, nil
}

func (path AbsolutePath) Value() string { return path.value }

func newAbsolutePath(value string) (AbsolutePath, bool) {
	if len(value) > 4096 || strings.IndexByte(value, 0) >= 0 ||
		!filepath.IsAbs(value) || filepath.Clean(value) != value {
		return AbsolutePath{}, false
	}
	return AbsolutePath{value: value}, true
}
