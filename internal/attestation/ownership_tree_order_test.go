package attestation

import (
	"bytes"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestVerifyOwnershipTreeEnforcesGlobalObjectOrder(t *testing.T) {
	root, sources := testOrderedOwnershipTree(t, false)
	if _, err := VerifyOwnershipTree(root, sources); err != nil {
		t.Fatalf("ordered ownership tree rejected: %v", err)
	}

	root, sources = testOrderedOwnershipTree(t, true)
	if _, err := VerifyOwnershipTree(root, sources); err == nil {
		t.Fatal("cross-chunk ownership order drift accepted")
	}
}

func testOrderedOwnershipTree(
	t *testing.T,
	reverse bool,
) (VerifiedPostgresArtifact[postgresv1.OwnershipManifest], []OwnershipChunkSource) {
	t.Helper()
	objects := []postgresv1.OwnershipObject{testOwnershipObject(), testSchemaOwnershipObject()}
	if reverse {
		objects[0], objects[1] = objects[1], objects[0]
	}
	descriptors := make([]postgresv1.OwnershipChunkDescriptor, 0, 2)
	sources := make([]OwnershipChunkSource, 0, 2)
	for index, object := range objects {
		chunk := postgresv1.OwnershipChunk{
			ContractVersion: "postgres-ownership-chunk.v1",
			ManifestID:      "11111111-1111-4111-8111-111111111111",
			DatabaseOrigin:  "fresh", DatabaseName: "harpy",
			Ordinal: index, Count: 2, ObjectCount: 1,
			Objects: []postgresv1.OwnershipObject{object},
		}
		document := testDocument(t, chunk)
		canonical, err := canonicalizeBounded(document, postgresv1.MaximumOwnershipChunkBytes)
		if err != nil {
			t.Fatal(err)
		}
		digest, err := DigestHex(document)
		if err != nil {
			t.Fatal(err)
		}
		path := "ownership/" + digest + ".json"
		descriptors = append(descriptors, postgresv1.OwnershipChunkDescriptor{
			Ordinal: index, Count: 2, LengthBytes: int64(len(canonical)),
			SHA256: digest, Path: path,
		})
		sources = append(sources, OwnershipChunkSource{
			Path: path, Reader: bytes.NewReader(document),
		})
	}
	manifest := postgresv1.OwnershipManifest{
		ContractVersion: "postgres-ownership-manifest.v1",
		ManifestID:      "11111111-1111-4111-8111-111111111111",
		DatabaseOrigin:  "fresh", DatabaseName: "harpy",
		OwnershipExecutor: "cluster_bootstrap", Format: "chunked", ObjectCount: 2,
		Chunks: descriptors,
	}
	verified, err := VerifyOwnershipManifestArtifact(testDocument(t, manifest))
	if err != nil {
		t.Fatal(err)
	}
	return verified, sources
}

func testSchemaOwnershipObject() postgresv1.OwnershipObject {
	return postgresv1.OwnershipObject{
		Object: postgresv1.CatalogObject{
			Kind: "schema", DatabaseName: "harpy", ObjectName: "platform",
		},
		Owner: "harpy_owner", Classification: "first_party", Executor: "cluster_bootstrap",
	}
}
