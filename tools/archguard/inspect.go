package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func inspect(root string) ([]string, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	root = filepath.Clean(absolute)
	module, err := readModulePath(root)
	if err != nil {
		return nil, err
	}
	previous := modulePath
	modulePath = module
	defer func() { modulePath = previous }()
	var violations []string
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && skipTree(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		sourcePath := filepath.ToSlash(relative)
		if message := removedTree(sourcePath); message != "" {
			violations = append(violations, sourcePath+": "+message)
		}
		if !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		source := classifySource(sourcePath)
		imports, parseErr := fileImports(path)
		if parseErr != nil {
			violations = append(violations, fmt.Sprintf("%s: cannot parse imports: %v", sourcePath, parseErr))
			return nil
		}
		for _, importPath := range imports {
			if violation := importViolation(source, importPath); violation != "" {
				violations = append(violations,
					fmt.Sprintf("%s: import %q: %s", sourcePath, importPath, violation))
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(violations)
	return violations, nil
}

func readModulePath(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return "", fmt.Errorf("read go.mod: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(strings.TrimRight(line, "\r"))
		if !strings.HasPrefix(line, "module ") {
			continue
		}
		module := strings.TrimSpace(strings.TrimPrefix(line, "module "))
		module = strings.Trim(module, `"`)
		if module == "" || strings.ContainsAny(module, " \t") {
			return "", fmt.Errorf("go.mod module path is invalid")
		}
		return module, nil
	}
	return "", fmt.Errorf("go.mod does not contain a module path")
}

func skipTree(name string) bool {
	return strings.HasPrefix(name, ".") || name == "vendor" ||
		name == "node_modules" || name == "testdata"
}

func fileImports(filename string) ([]string, error) {
	parsed, err := parser.ParseFile(token.NewFileSet(), filename, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, err
	}
	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		if spec.Path == nil {
			return nil, fmt.Errorf("invalid import literal")
		}
		value, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid import literal")
		}
		imports = append(imports, value)
	}
	return imports, nil
}
