package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const envLocalFileName = ".env.local"

const envLocalContents = `COMPOSE_PROJECT_NAME=harpy-local
LOCAL_API_PORT=8000
LOCAL_API_MANAGEMENT_PORT=9001
LOCAL_WORKER_MANAGEMENT_PORT=9002
LOCAL_SCHEMA_MANAGEMENT_PORT=9003
LOCAL_TENANT_MIGRATION_MANAGEMENT_PORT=9004
LOCAL_POSTGRES_PORT=55432
LOCAL_RABBITMQ_PORT=5673
LOCAL_RABBITMQ_MANAGEMENT_PORT=15673
LOCAL_S3_PORT=19000
LOCAL_MAILPIT_SMTP_PORT=1026
LOCAL_MAILPIT_UI_PORT=8026
LOCAL_PUBLIC_BASE_URL=http://localhost:5173
`

func ensureEnvLocal() error {
	start, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("read working directory: %w", err)
	}
	return ensureEnvLocalFrom(start)
}

func ensureEnvLocalFrom(start string) error {
	root, err := moduleRoot(start)
	if err != nil {
		return err
	}
	return writeEnvLocalIfAbsent(filepath.Join(root, envLocalFileName))
}

func moduleRoot(dir string) (string, error) {
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("module root is unavailable")
		}
		dir = parent
	}
}

func writeEnvLocalIfAbsent(path string) error {
	if _, err := os.Lstat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("inspect .env.local: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return fmt.Errorf("create .env.local: %w", err)
	}
	_, writeErr := file.WriteString(envLocalContents)
	var syncErr error
	if writeErr == nil {
		syncErr = file.Sync()
	}
	closeErr := file.Close()
	if writeErr != nil || syncErr != nil {
		_ = os.Remove(path)
	}
	if writeErr != nil {
		return fmt.Errorf("write .env.local: %w", writeErr)
	}
	if syncErr != nil {
		return fmt.Errorf("sync .env.local: %w", syncErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close .env.local: %w", closeErr)
	}
	return nil
}
