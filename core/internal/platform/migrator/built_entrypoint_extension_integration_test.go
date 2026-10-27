//go:build integration

package migrator

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func builtEntrypointExtensionDigest(t *testing.T) string {
	t.Helper()
	document, err := attestation.MarshalCanonical(builtEntrypointExtensionManifest())
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:])
}

func builtEntrypointExtensionManifest() postgresv1.ExtensionManifest {
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
				ControlFileSHA256: integrationDigest("pg-trgm-control"),
				Capabilities:      []string{"trigram_search"}},
			{Name: "vector", Version: "0.8.3", Install: true, Schema: "platform_ext",
				Owner: "harpy_extension_admin", ControlFile: "vector.control",
				ControlFileSHA256: integrationDigest("vector-control"),
				Capabilities:      []string{"vector_similarity"}},
		},
	}
}
