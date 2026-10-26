package deployevidence

import (
	"fmt"
	"path/filepath"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func loadOwnershipTree(path string) (attestation.VerifiedOwnershipTree, error) {
	document, err := readArtifact(
		path, postgresv1.MaximumOwnershipManifestBytes, "ownership descriptor",
	)
	if err != nil {
		return attestation.VerifiedOwnershipTree{}, err
	}
	root, err := attestation.VerifyOwnershipManifestArtifact(document.bytes)
	if err != nil {
		return attestation.VerifiedOwnershipTree{}, fmt.Errorf("ownership descriptor verification failed")
	}
	manifest := root.Value()
	if manifest.Format == "inline" {
		tree, err := attestation.VerifyOwnershipTree(root, nil)
		if err != nil {
			return attestation.VerifiedOwnershipTree{}, fmt.Errorf("ownership tree verification failed")
		}
		return tree, nil
	}
	baseDirectory := filepath.Dir(document.resolvedPath)
	sources := make([]attestation.OwnershipChunkSource, 0, len(manifest.Chunks))
	for _, descriptor := range manifest.Chunks {
		reader := &ownershipChunkReader{
			baseDirectory: baseDirectory, relativePath: descriptor.Path,
			maximumBytes: postgresv1.MaximumOwnershipChunkBytes,
		}
		sources = append(sources, attestation.OwnershipChunkSource{
			Path: descriptor.Path, Reader: reader,
		})
	}
	tree, err := attestation.VerifyOwnershipTree(root, sources)
	if err != nil {
		return attestation.VerifiedOwnershipTree{}, fmt.Errorf("ownership tree verification failed")
	}
	return tree, nil
}
