package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Onurlulardan/harpy/internal/migrationbundle"
)

func main() {
	check := flag.Bool("check", false, "compare release filenames with the index")
	output := flag.String("output", "", "zip path")
	rootFlag := flag.String("root", "", "module root")
	flag.Parse()
	root, err := moduleRoot(*rootFlag)
	if err != nil {
		fail(err)
	}
	if *check {
		if err = migrationbundle.ReleaseNamesMatchIndex(root); err != nil {
			fail(err)
		}
		fmt.Println("release archive names match the index")
		return
	}
	if *output == "" {
		fail(fmt.Errorf("output path is required"))
	}
	if err = migrationbundle.PackRelease(root, *output); err != nil {
		fail(err)
	}
	fmt.Printf("wrote %s\n", *output)
}

func moduleRoot(root string) (string, error) {
	if root != "" {
		return root, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
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

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
