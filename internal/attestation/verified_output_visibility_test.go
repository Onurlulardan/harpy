package attestation

import (
	"reflect"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestVerifiedOutputsDoNotExposeForgeableFields(t *testing.T) {
	outputs := []any{
		VerifiedRelease{},
		VerifiedBootstrapAuthority{},
		VerifiedBootstrapReceipt{},
		VerifiedExtensionReceipt{},
		VerifiedCredentialCleanupReceipt{},
		VerifiedCutoverCredentialLease{},
		VerifiedCutoverSessionSeal{},
		VerifiedCutoverCredentialCleanup{},
		VerifiedPostgresArtifact[postgresv1.MigrationManifest]{},
		VerifiedOwnershipTree{},
		VerifiedPostgresArtifactSet{},
	}
	for _, output := range outputs {
		outputType := reflect.TypeOf(output)
		t.Run(outputType.Name(), func(t *testing.T) {
			for index := 0; index < outputType.NumField(); index++ {
				field := outputType.Field(index)
				if field.IsExported() {
					t.Fatalf("verified output exposes forgeable field %q", field.Name)
				}
			}
		})
	}
}

func TestVerifiedReleaseReturnsDefensiveValue(t *testing.T) {
	manifest, signature, root := releaseTestDocuments(t)
	verified, err := VerifyRelease(manifest, signature, root, releaseTestNow, "harpy")
	if err != nil {
		t.Fatal(err)
	}
	value := verified.Value()
	value.DatabaseName = "tampered"
	if verified.Value().DatabaseName != "harpy" {
		t.Fatal("verified release exposed mutable manifest storage")
	}
	if (VerifiedRelease{}).Value() != (ReleaseManifest{}) ||
		(VerifiedRelease{}).ManifestSHA256() != "" ||
		!(VerifiedRelease{}).IssuedAt().IsZero() {
		t.Fatal("unverified release exposed a value")
	}
}
