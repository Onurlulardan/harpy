package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const migrationRoot = "core/internal/platform/goose/sql"

func main() {
	base := "origin/cursor/migration-initialize-foundation-b30d"
	if len(os.Args) == 2 {
		base = os.Args[1]
	}
	if err := check(base); err != nil {
		fmt.Fprintln(os.Stderr, "gooseimmutable:", err)
		os.Exit(1)
	}
	fmt.Println("gooseimmutable: OK")
}

func check(base string) error {
	previous, err := gitFiles(base)
	if err != nil {
		return err
	}
	current, err := worktreeFiles()
	if err != nil {
		return err
	}
	return changedMigrations(previous, current)
}

func worktreeFiles() (map[string]string, error) {
	files := map[string]string{}
	err := filepath.WalkDir(migrationRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() || !strings.HasSuffix(path, ".sql") {
			return walkErr
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(path)] = string(body)
		return nil
	})
	return files, err
}

func gitFiles(base string) (map[string]string, error) {
	command := exec.Command("git", "ls-tree", "-r", "--name-only", base, migrationRoot)
	output, err := command.Output()
	if err != nil {
		if len(output) == 0 {
			return map[string]string{}, nil
		}
		return nil, err
	}
	files := map[string]string{}
	for _, path := range strings.Fields(string(output)) {
		show := exec.Command("git", "show", base+":"+path)
		body, err := show.Output()
		if err != nil {
			return nil, err
		}
		files[path] = string(body)
	}
	return files, nil
}
