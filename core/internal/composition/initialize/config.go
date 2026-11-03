package initialize

import (
	"fmt"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/platform/foundation"
)

func loadConfig(lookup func(string) string, read func(string) ([]byte, error)) (foundation.Config, error) {
	admin, err := readSecret(lookup("HARPY_FOUNDATION_ADMIN_DSN_FILE"), read)
	if err != nil {
		return foundation.Config{}, fmt.Errorf("read foundation admin dsn: %w", err)
	}
	goose, err := readSecret(lookup("HARPY_GOOSE_PASSWORD_FILE"), read)
	if err != nil {
		return foundation.Config{}, fmt.Errorf("read goose password: %w", err)
	}
	api, err := readSecret(lookup("HARPY_API_PASSWORD_FILE"), read)
	if err != nil {
		return foundation.Config{}, fmt.Errorf("read api password: %w", err)
	}
	worker, err := readSecret(lookup("HARPY_WORKER_PASSWORD_FILE"), read)
	if err != nil {
		return foundation.Config{}, fmt.Errorf("read worker password: %w", err)
	}
	return foundation.Config{
		AdminDSN: admin, Database: lookup("HARPY_FOUNDATION_DATABASE"),
		GoosePassword: goose, APIPassword: api, WorkerPassword: worker,
	}, nil
}

func readSecret(path string, read func(string) ([]byte, error)) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}
	content, err := read(path)
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(content))
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return "", fmt.Errorf("secret file must contain one value")
	}
	return value, nil
}
