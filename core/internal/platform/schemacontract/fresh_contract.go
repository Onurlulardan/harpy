package schemacontract

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

// Control-file hashes are the PostgreSQL 18.4 image hashes recorded by
// backend-v2 localrelease. This tree does not vendor the control files.
const (
	pgTrgmControlFileSHA256 = "17c2745da2a706e1443ee363ab7424ced871a9a652e86bf55d25d139f43ee27b"
	vectorControlFileSHA256 = "db4ad0045e6622228471acdcb8fc9aeeec0cd6cf9666569d26f723751ea95fcb"
)

func freshExtensionDigest() (string, error) {
	document, err := attestation.MarshalCanonical(freshExtensionManifest())
	if err != nil {
		return "", fmt.Errorf("encode fresh extension manifest: %w", err)
	}
	artifact, err := attestation.VerifyExtensionManifestArtifact(document)
	if err != nil {
		return "", fmt.Errorf("verify fresh extension manifest: %w", err)
	}
	return artifact.DigestSHA256(), nil
}

func freshExtensionManifest() postgresv1.ExtensionManifest {
	return postgresv1.ExtensionManifest{
		ContractVersion: "postgres-extension-manifest.v1", DatabaseOrigin: "fresh",
		PostgresImageDigest:   postgresv1.PostgresImageDigest,
		PostgresEngineVersion: postgresv1.PostgresEngineVersion,
		ControlPathPolicy: postgresv1.ControlPathPolicy{
			Resolver: "pg_config_sharedir", RelativeDirectory: "extension",
		},
		Extensions: []postgresv1.PostgresExtension{
			{Name: "pg_trgm", Version: "1.6", Install: true, Schema: "platform_ext",
				Owner: "harpy_extension_admin", ControlFile: "pg_trgm.control",
				ControlFileSHA256: pgTrgmControlFileSHA256,
				Capabilities:      []string{"trigram_search"}},
			{Name: "vector", Version: "0.8.3", Install: true, Schema: "platform_ext",
				Owner: "harpy_extension_admin", ControlFile: "vector.control",
				ControlFileSHA256: vectorControlFileSHA256,
				Capabilities:      []string{"vector_similarity"}},
		},
	}
}
