package schemacontract

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/binarycontract"
)

func LoadBoundStartup() (EmbeddedContract, error) {
	return loadBoundStartup(
		runtimev1.ReadEmbeddedFreshMigrationManifest,
		binarycontract.LoadEmbedded,
	)
}

func loadBoundStartup(
	readManifest func() ([]byte, error),
	loadContract func() (binarycontract.Artifact, error),
) (EmbeddedContract, error) {
	document, manifestErr := readManifest()
	artifact, contractErr := loadContract()
	if contractErr != nil && manifestErr != nil {
		return EmbeddedContract{}, fmt.Errorf("%w; %v", contractErr, manifestErr)
	}
	if contractErr != nil {
		return EmbeddedContract{}, contractErr
	}
	if manifestErr != nil {
		return EmbeddedContract{}, manifestErr
	}
	return authorizeBoundManifest(document, artifact)
}

func authorizeBoundManifest(
	document []byte, artifact binarycontract.Artifact,
) (EmbeddedContract, error) {
	if !artifact.IsVerified() {
		return EmbeddedContract{}, fmt.Errorf("embedded binary schema contract is not verified")
	}
	if len(document) == 0 {
		return EmbeddedContract{}, fmt.Errorf("fresh migration manifest is empty")
	}
	sum := sha256.Sum256(document)
	if artifact.Value().MigrationManifestSHA256 != hex.EncodeToString(sum[:]) {
		return EmbeddedContract{}, fmt.Errorf(
			"schema contract migration digest does not match the fresh migration bundle",
		)
	}
	return EmbeddedContract{artifact: artifact, authorized: true}, nil
}
