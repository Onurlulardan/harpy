package migrationmanifest

import (
	"encoding/json"
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func Encode(manifest postgresv1.MigrationManifest) ([]byte, string, error) {
	encoded, err := json.Marshal(manifest)
	if err != nil {
		return nil, "", fmt.Errorf("encode migration manifest: %w", err)
	}
	verified, err := attestation.VerifyMigrationManifestArtifact(encoded)
	if err != nil {
		return nil, "", fmt.Errorf("verify generated migration manifest: %w", err)
	}
	return verified.CanonicalJSON(), verified.DigestSHA256(), nil
}
