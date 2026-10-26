package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
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
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		sourcePath := filepath.ToSlash(relative)
		source := classifySource(sourcePath)
		if message := removedTree(sourcePath); message != "" {
			violations = append(violations, sourcePath+": "+message)
		}
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
