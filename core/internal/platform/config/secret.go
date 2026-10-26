package config

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
)

const secretFileLimit = 64 * 1024

type Secret struct {
	value string
}

func (s Secret) Value() string {
	return s.value
}

func (Secret) String() string {
	return "[REDACTED]"
}

func (Secret) GoString() string {
	return "[REDACTED]"
}

func (Secret) LogValue() slog.Value {
	return slog.StringValue("[REDACTED]")
}

func loadSecret(values environmentValues, key string, required bool, maximum int) (Secret, error) {
	direct, directSet := values.lookup(key)
	filePath, fileSet := values.lookup(key + "_FILE")
	if directSet && fileSet {
		return Secret{}, configError(key, "value and file are mutually exclusive")
	}
	if !directSet && !fileSet {
		if required {
			return Secret{}, configError(key, "is required")
		}
		return Secret{}, nil
	}
	value := direct
	if fileSet {
		loaded, err := readSecretFile(key, filePath, maximum)
		if err != nil {
			return Secret{}, err
		}
		value = loaded
	}
	if err := validateSecretValue(key, value, maximum); err != nil {
		return Secret{}, err
	}
	return Secret{value: value}, nil
}

func readSecretFile(key, path string, maximum int) (string, error) {
	if path == "" || !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return "", configError(key, "secret file path must be absolute and clean")
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", configError(key, "secret file cannot be resolved")
	}
	file, err := os.Open(resolved)
	if err != nil {
		return "", configError(key, "secret file cannot be opened")
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return "", configError(key, "secret file must resolve to a regular file")
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o022 != 0 {
		return "", configError(key, "secret file permissions are too broad")
	}
	limit := min(maximum, secretFileLimit)
	content, err := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	if err != nil || len(content) > limit {
		return "", configError(key, "secret file exceeds its size limit")
	}
	content = removeOneTerminalNewline(content)
	return string(content), nil
}

func removeOneTerminalNewline(content []byte) []byte {
	if bytes.HasSuffix(content, []byte("\r\n")) {
		return content[:len(content)-2]
	}
	if bytes.HasSuffix(content, []byte("\n")) {
		return content[:len(content)-1]
	}
	return content
}

func validateSecretValue(key, value string, maximum int) error {
	if value == "" {
		return configError(key, "must not be empty")
	}
	if len(value) > maximum {
		return configError(key, "exceeds its size limit")
	}
	if bytes.IndexByte([]byte(value), 0) >= 0 {
		return configError(key, "contains a prohibited byte")
	}
	return nil
}

func (s Secret) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED]")
}
