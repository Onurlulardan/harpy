package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	localFrontendReadyURL       = "http://localhost:5173/"
	legacyLocalFrontendReadyURL = "http://127.0.0.1:5173/"
)

type runtimeConfig struct {
	BackendRoot   string
	FrontendRoot  string
	StateDir      string
	ReceiptPath   string
	LogPath       string
	ReadyURL      string
	LocalAPIPort  string
	StartupWindow time.Duration
}

func loadConfig() (runtimeConfig, error) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return runtimeConfig{}, fmt.Errorf("read working directory: %w", err)
	}
	backendRoot, err := findBackendRoot(workingDirectory)
	if err != nil {
		return runtimeConfig{}, err
	}
	frontendRoot := filepath.Join(filepath.Dir(backendRoot), "frontend-v3")
	for _, required := range []string{"package.json", "yarn.lock"} {
		if _, err := os.Stat(filepath.Join(frontendRoot, required)); err != nil {
			return runtimeConfig{}, fmt.Errorf("frontend-v3 %s is unavailable: %w", required, err)
		}
	}
	apiPort, err := localAPIPort(filepath.Join(backendRoot, ".env.local"))
	if err != nil {
		return runtimeConfig{}, err
	}
	stateDir := filepath.Join(backendRoot, "ops", "local", "state", "frontend-v3")
	return runtimeConfig{
		BackendRoot:   backendRoot,
		FrontendRoot:  frontendRoot,
		StateDir:      stateDir,
		ReceiptPath:   filepath.Join(stateDir, "receipt.json"),
		LogPath:       filepath.Join(stateDir, "frontend.log"),
		ReadyURL:      localFrontendReadyURL,
		LocalAPIPort:  apiPort,
		StartupWindow: 60 * time.Second,
	}, nil
}

func findBackendRoot(start string) (string, error) {
	current := filepath.Clean(start)
	for {
		modulePath := filepath.Join(current, "go.mod")
		content, err := os.ReadFile(modulePath)
		if err == nil && strings.Contains(string(content), "module github.com/Onurlulardan/harpy") {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("backend-v2 root was not found from %s", start)
		}
		current = parent
	}
}

func localAPIPort(path string) (string, error) {
	value, found, err := readEnvValue(path, "LOCAL_API_PORT")
	if err != nil {
		return "", err
	}
	if !found || value == "" {
		value = "8000"
	}
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return "", fmt.Errorf("LOCAL_API_PORT must be a valid TCP port, received %q", value)
	}
	return strconv.Itoa(port), nil
}

func readEnvValue(path, key string) (string, bool, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, value, found := strings.Cut(line, "=")
		if found && strings.TrimSpace(name) == key {
			return strings.Trim(strings.TrimSpace(value), "\"'"), true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", false, fmt.Errorf("read %s: %w", path, err)
	}
	return "", false, nil
}

func setEnvironment(environment []string, key, value string) []string {
	prefix := strings.ToUpper(key) + "="
	result := make([]string, 0, len(environment)+1)
	for _, entry := range environment {
		if !strings.HasPrefix(strings.ToUpper(entry), prefix) {
			result = append(result, entry)
		}
	}
	return append(result, key+"="+value)
}
