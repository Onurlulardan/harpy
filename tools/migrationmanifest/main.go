package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/Onurlulardan/harpy/tools/internal/migrationmanifest"
)

const defaultArtifactDirectory = "core/internal/platform/migration/release"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "migrationmanifest:", err)
		os.Exit(1)
	}
}

func run() error {
	rootFlag := flag.String("root", ".", "repository root")
	directoryFlag := flag.String(
		"artifact-dir", defaultArtifactDirectory, "immutable artifact directory",
	)
	write := flag.Bool("write", false, "publish missing immutable artifact")
	markApplied := flag.Bool(
		"mark-applied", false, "clear unpublished marker after successful deploy",
	)
	flag.Parse()
	root, err := filepath.Abs(*rootFlag)
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}
	inventory, err := migrationmanifest.LoadStorageInventory(root)
	if err != nil {
		return err
	}
	groups, err := migrationmanifest.CollectAll(root, inventory)
	if err != nil {
		return err
	}
	manifest, err := migrationmanifest.Build(inventory, groups)
	if err != nil {
		return err
	}
	directory, err := resolveArtifactDirectory(root, *directoryFlag)
	if err != nil {
		return err
	}
	validate := migrationmanifest.ValidateReleaseFrontierAdvance
	if *write {
		validate = migrationmanifest.ValidateReleaseFrontierAdvanceForPublish
	}
	if err := validate(directory, manifest); err != nil {
		return err
	}
	canonical, digest, err := migrationmanifest.Encode(manifest)
	if err != nil {
		return err
	}
	target := filepath.Join(directory, digest+".migration-manifest.json")
	created, err := verifyOrPublish(target, canonical, *write)
	if err != nil {
		return err
	}
	if created {
		if err := migrationmanifest.WriteUnpublishedMarker(directory, digest); err != nil {
			return err
		}
	}
	if *markApplied {
		if err := migrationmanifest.MarkReleaseApplied(directory, digest); err != nil {
			return err
		}
	}
	fmt.Printf("migration manifest: %s\nsha256: %s\n", target, digest)
	return nil
}

func resolveArtifactDirectory(root, relative string) (string, error) {
	if !fs.ValidPath(relative) || filepath.IsAbs(relative) {
		return "", fmt.Errorf("artifact directory must be canonical and repository-relative")
	}
	target := filepath.Clean(filepath.Join(root, filepath.FromSlash(relative)))
	relativeTarget, err := filepath.Rel(root, target)
	if err != nil || relativeTarget == ".." ||
		len(relativeTarget) > 3 && relativeTarget[:3] == ".."+string(filepath.Separator) {
		return "", fmt.Errorf("artifact directory escapes repository root")
	}
	return target, nil
}

func verifyOrPublish(target string, canonical []byte, write bool) (bool, error) {
	existing, err := os.ReadFile(target)
	if err == nil {
		if !bytes.Equal(existing, canonical) {
			return false, fmt.Errorf("digest-named manifest content differs")
		}
		return false, nil
	}
	if !os.IsNotExist(err) {
		return false, fmt.Errorf("read migration manifest: %w", err)
	}
	if !write {
		return false, fmt.Errorf("migration manifest is missing; run with -write")
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return false, fmt.Errorf("create migration manifest directory: %w", err)
	}
	if err := attestation.WriteNewCanonicalBounded(
		target, canonical, postgresv1.MaximumMigrationManifestBytes,
	); err != nil {
		return false, err
	}
	return true, nil
}
