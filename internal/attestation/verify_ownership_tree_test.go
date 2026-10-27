package attestation

import (
	"bytes"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func TestVerifyOwnershipTreeBindsChunksAndCounts(t *testing.T) {
	chunk := testOwnershipChunk(1)
	document := testDocument(t, chunk)
	root, path := testChunkedRoot(t, document, 1)
	tree, err := VerifyOwnershipTree(root, []OwnershipChunkSource{{
		Path: path, Reader: bytes.NewReader(document),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if !tree.IsVerified() || tree.ObjectCount() != 1 || len(tree.ChunkDigests()) != 1 {
		t.Fatal("verified ownership tree evidence is incomplete")
	}
	digests := tree.ChunkDigests()
	digests[0] = "mutated"
	if tree.ChunkDigests()[0] == "mutated" {
		t.Fatal("ownership tree exposes mutable digest state")
	}
}

func TestVerifyOwnershipTreeRejectsHashLengthIdentityAndOversize(t *testing.T) {
	validChunk := testDocument(t, testOwnershipChunk(1))
	tests := []struct {
		name  string
		build func(*testing.T) (VerifiedPostgresArtifact[postgresv1.OwnershipManifest], string, []byte)
	}{
		{"wrong hash", func(t *testing.T) (VerifiedPostgresArtifact[postgresv1.OwnershipManifest], string, []byte) {
			root, path := testChunkedRootWith(t, validChunk, 1, func(root *postgresv1.OwnershipManifest) {
				root.Chunks[0].SHA256 = testHash("f")
				root.Chunks[0].Path = "ownership/" + testHash("f") + ".json"
			})
			return root, path, validChunk
		}},
		{"wrong canonical length", func(t *testing.T) (VerifiedPostgresArtifact[postgresv1.OwnershipManifest], string, []byte) {
			root, path := testChunkedRootWith(t, validChunk, 1, func(root *postgresv1.OwnershipManifest) {
				root.Chunks[0].LengthBytes++
			})
			return root, path, validChunk
		}},
		{"chunk count mismatch", func(t *testing.T) (VerifiedPostgresArtifact[postgresv1.OwnershipManifest], string, []byte) {
			document := testDocument(t, testOwnershipChunk(2))
			root, path := testChunkedRoot(t, document, 1)
			return root, path, document
		}},
		{"total object count", func(t *testing.T) (VerifiedPostgresArtifact[postgresv1.OwnershipManifest], string, []byte) {
			root, path := testChunkedRoot(t, validChunk, 2)
			return root, path, validChunk
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, path, document := test.build(t)
			if _, err := VerifyOwnershipTree(root, []OwnershipChunkSource{{
				Path: path, Reader: bytes.NewReader(document),
			}}); err == nil {
				t.Fatal("tampered ownership tree accepted")
			}
		})
	}

	root, path := testChunkedRoot(t, validChunk, 1)
	oversize := bytes.Repeat([]byte{' '}, postgresv1.MaximumOwnershipChunkBytes+1)
	if _, err := VerifyOwnershipTree(root, []OwnershipChunkSource{{
		Path: path, Reader: bytes.NewReader(oversize),
	}}); err == nil {
		t.Fatal("oversize ownership chunk accepted")
	}
}

func testOwnershipChunk(count int) postgresv1.OwnershipChunk {
	return postgresv1.OwnershipChunk{
		ContractVersion: "postgres-ownership-chunk.v1",
		ManifestID:      "11111111-1111-4111-8111-111111111111",
		DatabaseOrigin:  "fresh", DatabaseName: "harpy",
		Ordinal: 0, Count: count, ObjectCount: 1,
		Objects: []postgresv1.OwnershipObject{testOwnershipObject()},
	}
}

func testChunkedRoot(
	t *testing.T,
	chunkDocument []byte,
	objectCount int64,
) (VerifiedPostgresArtifact[postgresv1.OwnershipManifest], string) {
	return testChunkedRootWith(t, chunkDocument, objectCount, nil)
}

func testChunkedRootWith(
	t *testing.T,
	chunkDocument []byte,
	objectCount int64,
	mutate func(*postgresv1.OwnershipManifest),
) (VerifiedPostgresArtifact[postgresv1.OwnershipManifest], string) {
	t.Helper()
	canonical, err := canonicalizeBounded(chunkDocument, postgresv1.MaximumOwnershipChunkBytes)
	if err != nil {
		t.Fatal(err)
	}
	digest, err := DigestHex(chunkDocument)
	if err != nil {
		t.Fatal(err)
	}
	path := "ownership/" + digest + ".json"
	root := postgresv1.OwnershipManifest{
		ContractVersion: "postgres-ownership-manifest.v1",
		ManifestID:      "11111111-1111-4111-8111-111111111111",
		DatabaseOrigin:  "fresh", DatabaseName: "harpy",
		OwnershipExecutor: "cluster_bootstrap", Format: "chunked", ObjectCount: objectCount,
		Chunks: []postgresv1.OwnershipChunkDescriptor{{
			Ordinal: 0, Count: 1, LengthBytes: int64(len(canonical)), SHA256: digest, Path: path,
		}},
	}
	if mutate != nil {
		mutate(&root)
		path = root.Chunks[0].Path
	}
	verified, err := VerifyOwnershipManifestArtifact(testDocument(t, root))
	if err != nil {
		t.Fatal(err)
	}
	return verified, path
}
