package schemacontract

import (
	"errors"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
)

func TestLoadEmbeddedPropagatesAbsenceWithoutAuthority(t *testing.T) {
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		t.Fatal(err)
	}
	contract, err := loadEmbedded(registry, func() (binarycontract.Artifact, error) {
		return binarycontract.Artifact{}, runtimev1.BinarySchemaContractNotEmbeddedError{}
	})
	var missing runtimev1.BinarySchemaContractNotEmbeddedError
	if !errors.As(err, &missing) {
		t.Fatalf("expected missing contract error, got %v", err)
	}
	if contract.IsVerified() {
		t.Fatal("missing contract returned schema authority")
	}
}

func TestAuthorizeEmbeddedRequiresMigrationRegistryMembership(t *testing.T) {
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		t.Fatal(err)
	}
	digests := registry.Digests()
	if len(digests) == 0 {
		t.Fatal("migration registry is empty")
	}
	artifact := contractArtifact(t, digests[0], "extension-a")
	contract, err := authorizeEmbedded(registry, artifact)
	if err != nil {
		t.Fatalf("authorize embedded contract: %v", err)
	}
	if !contract.IsVerified() || contract.Value().MigrationManifestSHA256 != digests[0] {
		t.Fatalf("unexpected embedded authority: %+v", contract.Value())
	}
	unembedded := contractArtifact(t, contract.Value().ExtensionManifestSHA256, "extension-b")
	if _, err := authorizeEmbedded(registry, unembedded); err == nil {
		t.Fatal("unembedded migration digest was authorized")
	}
}
