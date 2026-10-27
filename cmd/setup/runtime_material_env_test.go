package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func loadStartupProfile(t *testing.T, startup string, lines []string) error {
	t.Helper()
	path := filepath.Join(t.TempDir(), startup+".env")
	body := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		return err
	}
	binary, err := configLoaderBinary(t)
	if err != nil {
		return err
	}
	command := exec.Command(binary, startup, path)
	command.Env = []string{}
	output, err := command.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text == "" {
			text = err.Error()
		}
		return fmt.Errorf("%s", text)
	}
	return nil
}

var (
	configLoaderOnce  sync.Once
	configLoaderPath  string
	configLoaderError error
)

func configLoaderBinary(t *testing.T) (string, error) {
	t.Helper()
	configLoaderOnce.Do(func() {
		configLoaderPath, configLoaderError = buildConfigLoader()
	})
	return configLoaderPath, configLoaderError
}

func buildConfigLoader() (string, error) {
	root, err := findModuleRoot()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, "core", "internal", "setupruntimeboot")
	if err = os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)
	source := filepath.Join(root, "cmd", "setup", "testdata", "configloader", "main.go")
	if err = copyFile(source, filepath.Join(dir, "main.go")); err != nil {
		return "", err
	}
	binary := filepath.Join(os.TempDir(), "harpy-setup-runtime-boot"+exeSuffix())
	command := exec.Command(goTool(), "build", "-mod=readonly", "-o", binary, ".")
	command.Dir = dir
	command.Env = goBuildEnv()
	output, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build config loader: %w\n%s", err, output)
	}
	return binary, nil
}

func copyFile(source, target string) error {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = io.Copy(output, input)
	return err
}

func goTool() string {
	name := "go"
	if runtime.GOOS == "windows" {
		name = "go.exe"
	}
	return filepath.Join(runtime.GOROOT(), "bin", name)
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func goBuildEnv() []string {
	env := os.Environ()
	return setEnv(setEnv(env, "GOPROXY", "off"), "GOTOOLCHAIN", "local")
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for index, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			env[index] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}
