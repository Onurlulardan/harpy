//go:build integration

package migrator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

type builtEntrypointArtifact struct {
	path     string
	digest   string
	document []byte
}

func writeBuiltEntrypointArtifact(
	t *testing.T,
	directory string,
	name string,
	value any,
) builtEntrypointArtifact {
	t.Helper()
	if validatable, ok := value.(interface{ Validate() error }); ok {
		if err := validatable.Validate(); err != nil {
			t.Fatalf("validate built entrypoint artifact %s: %v", name, err)
		}
	}
	document, err := attestation.MarshalCanonical(value)
	if err != nil {
		t.Fatalf("encode built entrypoint artifact %s: %v", name, err)
	}
	return writeBuiltEntrypointDocument(t, directory, name, document)
}

func writeBuiltEntrypointDocument(
	t *testing.T,
	directory string,
	name string,
	document []byte,
) builtEntrypointArtifact {
	t.Helper()
	digest, err := attestation.DigestHex(document)
	if err != nil {
		t.Fatalf("digest built entrypoint artifact %s: %v", name, err)
	}
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, document, 0o600); err != nil {
		t.Fatalf("write built entrypoint artifact %s: %v", name, err)
	}
	return builtEntrypointArtifact{path: path, digest: digest, document: document}
}

func readBuiltEntrypointMigrationArtifact(
	t *testing.T,
	directory string,
	stage builtEntrypointStage,
) builtEntrypointArtifact {
	t.Helper()
	name := stage.migrationDigest + ".migration-manifest.json"
	path := filepath.Join(
		stage.directory, "core", "internal", "platform", "migration", "release", name,
	)
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read staged migration artifact: %v", err)
	}
	verified, err := attestation.VerifyMigrationManifestArtifact(document)
	if err != nil {
		t.Fatalf("verify staged migration artifact: %v", err)
	}
	canonical := verified.CanonicalJSON()
	artifactPath := filepath.Join(directory, name)
	if err := os.WriteFile(artifactPath, canonical, 0o600); err != nil {
		t.Fatalf("write staged migration artifact: %v", err)
	}
	artifact := builtEntrypointArtifact{
		path: artifactPath, digest: verified.DigestSHA256(), document: canonical,
	}
	if artifact.digest != stage.migrationDigest {
		t.Fatalf("staged migration artifact digest = %s", artifact.digest)
	}
	return artifact
}
