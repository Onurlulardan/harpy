package envfile

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func Consume(args []string, profile string) ([]string, error) {
	path, remaining, err := splitEnvFile(args)
	if err != nil {
		return nil, err
	}
	if path == "" {
		path = os.Getenv("HARPY_ENV_FILE")
	}
	if path == "" {
		return remaining, nil
	}
	if err := applyFile(path, profile); err != nil {
		return nil, err
	}
	return remaining, nil
}

func splitEnvFile(args []string) (string, []string, error) {
	var path string
	remaining := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "--env-file":
			if path != "" {
				return "", nil, fmt.Errorf("config HARPY_ENV_FILE: is duplicated")
			}
			index++
			if index >= len(args) || args[index] == "" || strings.HasPrefix(args[index], "-") {
				return "", nil, fmt.Errorf("config HARPY_ENV_FILE: path is required")
			}
			path = args[index]
		case strings.HasPrefix(arg, "--env-file="):
			if path != "" {
				return "", nil, fmt.Errorf("config HARPY_ENV_FILE: is duplicated")
			}
			path = strings.TrimPrefix(arg, "--env-file=")
			if path == "" {
				return "", nil, fmt.Errorf("config HARPY_ENV_FILE: path is required")
			}
		default:
			remaining = append(remaining, arg)
		}
	}
	return path, remaining, nil
}

func applyFile(path, profile string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("config HARPY_ENV_FILE: cannot read %s", path)
	}
	defer file.Close()
	values, err := parseFile(file)
	if err != nil {
		return err
	}
	for key, value := range values {
		if err := applyKey(profile, key, value); err != nil {
			return err
		}
	}
	return nil
}

func parseFile(file *os.File) (map[string]string, error) {
	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(strings.TrimRight(scanner.Text(), "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) == "" || value == "" {
			continue
		}
		key = strings.TrimSpace(key)
		if _, exists := values[key]; exists {
			return nil, fmt.Errorf("config %s: is duplicated", key)
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("config HARPY_ENV_FILE: cannot read %s", file.Name())
	}
	return values, nil
}

func applyKey(profile, key, value string) error {
	if managedEnvironmentKey(key) {
		if !config.StartupProfilesRecognize(key) {
			return fmt.Errorf("config %s: profile contains an unknown key", key)
		}
		if !config.StartupProfileAllows(profile, key) {
			return nil
		}
	}
	if os.Getenv(key) != "" {
		return nil
	}
	return os.Setenv(key, value)
}

func managedEnvironmentKey(key string) bool {
	return key == "DEPLOYMENT_MODE" || key == "GOMEMLIMIT" ||
		strings.HasPrefix(strings.ToUpper(key), "HARPY_")
}
