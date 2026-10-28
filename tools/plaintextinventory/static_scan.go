package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func runStatic(arguments []string) (int, error) {
	flags := flag.NewFlagSet("static", flag.ContinueOnError)
	workspaceFlag := flags.String("workspace", "", "workspace root")
	if err := flags.Parse(arguments); err != nil || flags.NArg() != 0 || *workspaceFlag == "" {
		return 0, fmt.Errorf("static scan arguments are invalid")
	}
	workspace, err := filepath.Abs(*workspaceFlag)
	if err != nil || !workspaceShapeValid(workspace) {
		return 0, fmt.Errorf("workspace is invalid")
	}
	verified := 0
	for _, root := range []string{"backend-v2", "frontend-v3"} {
		err = filepath.WalkDir(filepath.Join(workspace, root),
			func(path string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				relative, relErr := filepath.Rel(workspace, path)
				if relErr != nil {
					return relErr
				}
				normalized := filepath.ToSlash(relative)
				if entry.IsDir() && ignoredStaticDirectory(normalized) {
					return filepath.SkipDir
				}
				kind := staticFileKind(normalized, entry)
				if kind == staticSkip {
					return nil
				}
				if err := inspectStaticFile(path, normalized, kind); err != nil {
					return err
				}
				verified++
				return nil
			})
		if err != nil {
			return verified, err
		}
	}
	return verified, nil
}

func inspectStaticFile(path, relative string, kind staticKind) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	line := 0
	for scanner.Scan() {
		line++
		if violation := staticLineViolation(scanner.Text(), kind); violation != "" {
			return fmt.Errorf("%s:%d contains forbidden %s", relative, line, violation)
		}
	}
	if err = scanner.Err(); err != nil {
		return fmt.Errorf("scan %s: %w", relative, err)
	}
	return nil
}

func workspaceShapeValid(root string) bool {
	for _, path := range []string{
		"backend-v2/go.mod", "frontend-v3/package.json",
	} {
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil || !info.Mode().IsRegular() {
			return false
		}
	}
	return true
}

func ignoredStaticDirectory(relative string) bool {
	name := filepath.Base(filepath.FromSlash(relative))
	if name == ".git" || name == "node_modules" || name == "dist" ||
		name == "build" || name == "coverage" || name == ".cache" {
		return true
	}
	return strings.HasPrefix(relative, "backend-v2/ops/characterization")
}
