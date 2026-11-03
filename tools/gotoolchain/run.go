package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func run(ctx context.Context, args []string) error {
	versionFlag := flag.NewFlagSet("gotoolchain", flag.ContinueOnError)
	versionFlag.SetOutput(os.Stderr)
	version := versionFlag.String("version", "", "Go patch release to pin, such as 1.27.3")
	if err := versionFlag.Parse(args); err != nil {
		return err
	}
	if versionFlag.NArg() != 0 {
		return fmt.Errorf("unexpected arguments")
	}
	next, err := parseRelease(*version)
	if err != nil {
		return err
	}
	root, err := moduleRoot()
	if err != nil {
		return err
	}
	current, err := currentRelease(root)
	if err != nil {
		return err
	}
	if sameRelease(current, next) {
		return fmt.Errorf("requested release is already pinned")
	}
	files, err := loadPinnedFiles(root, current)
	if err != nil {
		return err
	}
	digest, err := bookwormDigest(ctx, next)
	if err != nil {
		return err
	}
	rewritten, err := rewritePins(files, current, next, digest)
	if err != nil {
		return err
	}
	if err := writePins(root, rewritten); err != nil {
		return err
	}
	if err := verifyPins(ctx, root); err != nil {
		return err
	}
	fmt.Fprintf(os.Stdout, "pinned %s -> %s\n", current.Toolchain(), next.Toolchain())
	fmt.Fprintf(os.Stdout, "image golang:%s-bookworm@%s\n", next.String(), digest)
	fmt.Fprintln(os.Stdout, "verification passed")
	fmt.Fprintln(os.Stdout, "rebuild local and production images before task dev")
	return nil
}

func currentRelease(root string) (goRelease, error) {
	content, err := os.ReadFile(filepath.Join(root, "contracts/runtime/v1/identity.go"))
	if err != nil {
		return goRelease{}, fmt.Errorf("read runtime Go contract: %w", err)
	}
	const marker = `GoVersion        = "`
	start := strings.Index(string(content), marker)
	if start < 0 {
		return goRelease{}, fmt.Errorf("runtime Go contract has no GoVersion pin")
	}
	rest := string(content)[start+len(marker):]
	end := strings.IndexByte(rest, '"')
	if end < 0 || !strings.HasPrefix(rest[:end], "go") {
		return goRelease{}, fmt.Errorf("runtime Go contract has no GoVersion pin")
	}
	return parseRelease(strings.TrimPrefix(rest[:end], "go"))
}

func writePins(root string, files map[string]string) error {
	for path, content := range files {
		target := filepath.Join(root, filepath.FromSlash(path))
		info, err := os.Stat(target)
		if err != nil {
			return fmt.Errorf("stat %s: %w", path, err)
		}
		if err := os.WriteFile(target, []byte(content), info.Mode().Perm()); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

func verifyPins(ctx context.Context, root string) error {
	commands := [][]string{
		{"go", "test", "./tools/gotoolchain", "./contracts/runtime/v1",
			"./core/internal/platform/buildinfo"},
		{"go", "run", "./tools/fileguard"},
		{"go", "run", "./tools/archguard"},
	}
	for _, command := range commands {
		cmd := exec.CommandContext(ctx, command[0], command[1:]...)
		cmd.Dir = root
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s failed: %w\n%s", strings.Join(command, " "), err, output)
		}
	}
	return nil
}

func moduleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("find module root: %w", err)
	}
	for {
		content, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err == nil && strings.Contains(string(content), "module github.com/Onurlulardan/harpy") {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("harpy module root was not found")
		}
		dir = parent
	}
}
