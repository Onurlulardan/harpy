// Command fileguard checks authored files.
// Run from the harpy module: go run ./tools/fileguard
//
// Generated files pass only when contracts/generated.yaml records provenance.
// That file is JSON. A generated header without it fails.
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maxLines = 200

var forbiddenGoNames = map[string]bool{
	"crud.go": true, "handler.go": true, "service.go": true,
	"repository.go": true, "utils.go": true, "helpers.go": true,
	"common.go": true, "misc.go": true,
}

var ignoredDirectories = map[string]bool{
	".git": true, "vendor": true, "node_modules": true,
}

func main() {
	root := flag.String("root", ".", "repository root")
	flag.Parse()

	violations, err := inspect(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fileguard:", err)
		os.Exit(2)
	}
	if len(violations) == 0 {
		fmt.Println("fileguard: OK")
		return
	}
	for _, violation := range violations {
		fmt.Fprintln(os.Stderr, "fileguard:", violation)
	}
	os.Exit(1)
}

func inspect(root string) ([]string, error) {
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	root = filepath.Clean(absolute)
	generatedFiles, err := loadGeneratedManifest(root)
	if err != nil {
		return nil, err
	}
	var violations []string
	seenGenerated := map[string]bool{}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && ignoredDirectories[strings.ToLower(entry.Name())] {
				return filepath.SkipDir
			}
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(relative)
		if entry.Type()&os.ModeSymlink != 0 {
			violations = append(violations, key+" is a symlink; authored links are not allowed")
			return nil
		}
		base := strings.ToLower(entry.Name())
		if isLockfile(base) {
			return nil
		}
		if provenance, registered := generatedFiles[key]; registered && isGeneratedBinary(base) {
			if err := verifyGenerated(path, provenance); err != nil {
				violations = append(violations, fmt.Sprintf("%s generated provenance: %v", key, err))
				return nil
			}
			seenGenerated[key] = true
			return nil
		}
		text, err := isTextFile(path)
		if err != nil {
			return err
		}
		if !text {
			violations = append(violations, key+" is not valid UTF-8 text")
			return nil
		}
		generated, err := hasGeneratedHeader(path)
		if err != nil {
			return err
		}
		if generated {
			provenance, registered := generatedFiles[key]
			if !registered {
				violations = append(violations, key+" has generated header without provenance")
				return nil
			}
			if err := verifyGenerated(path, provenance); err != nil {
				violations = append(violations, fmt.Sprintf("%s generated provenance: %v", key, err))
				return nil
			}
			seenGenerated[key] = true
			return nil
		}
		lines, err := countLines(path)
		if err != nil {
			return err
		}
		if lines > maxLines {
			violations = append(violations,
				fmt.Sprintf("%s has %d lines; maximum is %d", key, lines, maxLines))
		}
		if filepath.Ext(base) == ".go" && forbiddenGoNames[base] {
			violations = append(violations, key+" uses a forbidden dumping-ground name")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for path := range generatedFiles {
		if !seenGenerated[path] {
			violations = append(violations, path+" is registered generated output but was not verified")
		}
	}
	sort.Strings(violations)
	return violations, nil
}
