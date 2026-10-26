package schemacontract

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func TestBindExpectedRequiresExactCanonicalReleaseProjection(t *testing.T) {
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		t.Fatal(err)
	}
	digest := registry.Digests()[0]
	artifact := contractArtifact(t, digest, "extension-a")
	embedded, err := authorizeEmbedded(registry, artifact)
	if err != nil {
		t.Fatal(err)
	}
	binding, err := bindExpected(embedded, artifact)
	if err != nil {
		t.Fatalf("bind exact release projection: %v", err)
	}
	if !binding.IsVerified() || !binding.Contract().IsVerified() {
		t.Fatal("exact release projection did not produce authority")
	}
	different := contractArtifact(t, digest, "extension-b")
	if _, err := bindExpected(embedded, different); err == nil {
		t.Fatal("different release projection was bound")
	}
	if (ReleaseBinding{}).Contract().IsVerified() {
		t.Fatal("zero release binding exposed authority")
	}
}
