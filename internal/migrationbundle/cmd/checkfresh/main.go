package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/Onurlulardan/harpy/internal/migrationbundle"
)

func main() {
	write := flag.Bool("write", false, "write the contract digest and fresh migration manifest")
	rootFlag := flag.String("root", "", "module root")
	flag.Parse()
	if err := checkFresh(*rootFlag, *write); err != nil {
		fail(err)
	}
}

func checkFresh(rootFlag string, write bool) error {
	root, err := moduleRoot(rootFlag)
	if err != nil {
		return err
	}
	body, err := migrationbundle.CanonicalFresh(root)
	if err != nil {
		return err
	}
	digest := migrationbundle.Digest(body)
	if write {
		if err = writeFreshArtifacts(root, digest, body); err != nil {
			return err
		}
	}
	recorded, err := contractDigest(root)
	if err != nil {
		return err
	}
	if digest != recorded {
		return fmt.Errorf("fresh migration digest does not match the binary schema contract")
	}
	embedded, err := os.ReadFile(manifestPath(root))
	if err != nil {
		return err
	}
	if !bytes.Equal(body, embedded) {
		return fmt.Errorf("fresh migration manifest is stale; run with -write")
	}
	if _, err = attestation.VerifyMigrationManifestArtifact(body); err != nil {
		return err
	}
	fmt.Printf("fresh migration manifest %s %d bytes\n", digest, len(body))
	return nil
}

func manifestPath(root string) string {
	return filepath.Join(root, "contracts", "runtime", "v1", "fresh-migration-manifest.json")
}

func contractPath(root string) string {
	return filepath.Join(root, "contracts", "runtime", "v1", "binary-schema-contract.json")
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
