package attestation

import (
	"crypto/subtle"
	"fmt"
	"io"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

type OwnershipChunkSource struct {
	Path   string
	Reader io.Reader
}

type VerifiedOwnershipTree struct {
	root         VerifiedPostgresArtifact[postgresv1.OwnershipManifest]
	chunkDigests []string
	objectCount  int64
	verified     bool
}

func (tree VerifiedOwnershipTree) Root() VerifiedPostgresArtifact[postgresv1.OwnershipManifest] {
	return tree.root
}

func (tree VerifiedOwnershipTree) ChunkDigests() []string {
	return append([]string(nil), tree.chunkDigests...)
}

func (tree VerifiedOwnershipTree) ObjectCount() int64 {
	return tree.objectCount
}

func (tree VerifiedOwnershipTree) IsVerified() bool {
	return tree.verified
}

func VerifyOwnershipTree(
	root VerifiedPostgresArtifact[postgresv1.OwnershipManifest],
	sources []OwnershipChunkSource,
) (VerifiedOwnershipTree, error) {
	if !root.IsVerified() {
		return VerifiedOwnershipTree{}, fmt.Errorf("ownership root is not verified")
	}
	manifest := root.Value()
	if manifest.Format == "inline" {
		if len(sources) != 0 {
			return VerifiedOwnershipTree{}, fmt.Errorf("inline ownership has chunk sources")
		}
		return VerifiedOwnershipTree{
			root: root, objectCount: manifest.ObjectCount, verified: true,
		}, nil
	}
	if len(sources) != len(manifest.Chunks) {
		return VerifiedOwnershipTree{}, fmt.Errorf("ownership chunk source count differs")
	}
	return verifyOwnershipChunks(root, manifest, sources)
}

func verifyOwnershipChunks(
	root VerifiedPostgresArtifact[postgresv1.OwnershipManifest],
	manifest postgresv1.OwnershipManifest,
	sources []OwnershipChunkSource,
) (VerifiedOwnershipTree, error) {
	digests := make([]string, 0, len(sources))
	var total int64
	previousKey := ""
	for index, source := range sources {
		descriptor := manifest.Chunks[index]
		if source.Path != descriptor.Path {
			return VerifiedOwnershipTree{}, fmt.Errorf("ownership chunk %d path differs", index)
		}
		document, err := readBoundedDocument(source.Reader, postgresv1.MaximumOwnershipChunkBytes)
		if err != nil {
			return VerifiedOwnershipTree{}, fmt.Errorf("read ownership chunk %d: %w", index, err)
		}
		artifact, err := verifyPostgresArtifact[postgresv1.OwnershipChunk](
			document, postgresv1.MaximumOwnershipChunkBytes,
		)
		if err != nil {
			return VerifiedOwnershipTree{}, fmt.Errorf("verify ownership chunk %d: %w", index, err)
		}
		if int64(len(artifact.canonical)) != descriptor.LengthBytes ||
			subtle.ConstantTimeCompare(
				[]byte(artifact.DigestSHA256()), []byte(descriptor.SHA256),
			) != 1 {
			return VerifiedOwnershipTree{}, fmt.Errorf("ownership chunk %d content binding differs", index)
		}
		chunk := artifact.Value()
		if err := validateChunkBinding(manifest, descriptor, chunk); err != nil {
			return VerifiedOwnershipTree{}, fmt.Errorf("ownership chunk %d: %w", index, err)
		}
		for _, object := range chunk.Objects {
			key := postgresv1.OwnershipObjectKey(object)
			if previousKey != "" && previousKey >= key {
				return VerifiedOwnershipTree{}, fmt.Errorf("ownership tree is not globally ordered")
			}
			previousKey = key
		}
		total += chunk.ObjectCount
		digests = append(digests, artifact.DigestSHA256())
	}
	if total != manifest.ObjectCount {
		return VerifiedOwnershipTree{}, fmt.Errorf("ownership tree object count differs")
	}
	return VerifiedOwnershipTree{
		root: root, chunkDigests: digests, objectCount: total, verified: true,
	}, nil
}

func validateChunkBinding(
	manifest postgresv1.OwnershipManifest,
	descriptor postgresv1.OwnershipChunkDescriptor,
	chunk postgresv1.OwnershipChunk,
) error {
	if chunk.ManifestID != manifest.ManifestID ||
		chunk.DatabaseOrigin != manifest.DatabaseOrigin ||
		chunk.DatabaseName != manifest.DatabaseName ||
		chunk.Ordinal != descriptor.Ordinal || chunk.Count != descriptor.Count {
		return fmt.Errorf("identity differs from ownership root")
	}
	return nil
}
