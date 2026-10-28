package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/Onurlulardan/harpy/internal/migrationbundle"
)

func main() {
	write := flag.Bool("write", false, "write the fresh migration manifest")
	rootFlag := flag.String("root", "", "module root")
	flag.Parse()
	root, err := moduleRoot(*rootFlag)
	if err != nil {
		fail(err)
	}
	body, err := migrationbundle.CanonicalFresh(root)
	if err != nil {
		fail(err)
	}
	if migrationbundle.Digest(body) != contractDigest(root) {
		fail(fmt.Errorf("fresh migration digest does not match the binary schema contract"))
	}
	target := filepath.Join(root, "contracts", "runtime", "v1", "fresh-migration-manifest.json")
	if *write {
		if err = os.WriteFile(target, body, 0o644); err != nil {
			fail(err)
		}
	}
	embedded, err := os.ReadFile(target)
	if err != nil {
		fail(err)
	}
	if !bytes.Equal(body, embedded) {
		fail(fmt.Errorf("fresh migration manifest is stale; run with -write"))
	}
	if _, err = attestation.VerifyMigrationManifestArtifact(body); err != nil {
		fail(err)
	}
	fmt.Printf("fresh migration manifest %s %d bytes\n", migrationbundle.Digest(body), len(body))
}

func contractDigest(root string) string {
	document, err := os.ReadFile(filepath.Join(
		root, "contracts", "runtime", "v1", "binary-schema-contract.json",
	))
	if err != nil {
		fail(err)
	}
	contract, err := attestation.VerifyBinarySchemaContract(document)
	if err != nil {
		fail(err)
	}
	if contract.Value().ContractVersion != runtimev1.BinarySchemaContractVersion {
		fail(fmt.Errorf("binary schema contract version differs"))
	}
	return contract.Value().MigrationManifestSHA256
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
