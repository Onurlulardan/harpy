//go:build integration

package migrator

import (
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

const (
	builtEntrypointReleaseID      = "22222222-2222-4222-8222-222222222222"
	builtEntrypointEnvironmentID  = "44444444-4444-4444-8444-444444444444"
	builtEntrypointClusterLease   = "55555555-5555-4555-8555-555555555555"
	builtEntrypointExtensionLease = "66666666-6666-4666-8666-666666666666"
)

func builtEntrypointAuthorityManifest(
	now time.Time,
	catalog builtEntrypointArtifactCatalog,
) runtimev1.BootstrapAuthorityManifest {
	issuedAt := now.Add(-time.Minute)
	expiresAt := now.Add(9 * time.Minute)
	template := "template0"
	return runtimev1.BootstrapAuthorityManifest{
		ContractVersion: "bootstrap-authority-manifest.v1",
		ManifestID:      "33333333-3333-4333-8333-333333333334",
		EnvironmentID:   builtEntrypointEnvironmentID, ReleaseID: builtEntrypointReleaseID,
		IssuedAt: issuedAt.Format(time.RFC3339), ExpiresAt: expiresAt.Format(time.RFC3339),
		DatabaseOrigin: "fresh",
		Database: runtimev1.BootstrapAuthorityDatabase{
			Name: builtEntrypointDatabaseName, CreateDatabase: true,
			Owner: "harpy_owner", ExtensionOwner: "harpy_extension_admin",
			ExtensionOwnerMode: "fresh_leased_admin", OwnershipExecutor: "cluster_bootstrap",
			OwnershipAttestation: "bootstrap_live_catalog",
			Encoding:             runtimev1.PostgresEncoding,
			LocaleProvider:       runtimev1.PostgresLocaleProvider,
			Collation:            runtimev1.PostgresCollation, CType: runtimev1.PostgresCharacterType,
			DataChecksums: runtimev1.PostgresDataChecksums, Template: &template,
		},
		Connections: runtimev1.BootstrapAuthorityConnections{
			EndpointKind: "direct_postgres", MaintenanceDatabase: "postgres",
			TargetDatabase: builtEntrypointDatabaseName, SameServerRequired: true,
		},
		Roles: builtEntrypointAuthorityRoles(),
		Memberships: []runtimev1.BootstrapAuthorityMembership{
			builtEntrypointAuthorityMembership(
				"harpy_v2_runtime", "harpy_runtime", true, false, true,
			),
			builtEntrypointAuthorityMembership(
				"harpy_v2_runtime", "harpy_forms_record_reader", true, false, true,
			),
			builtEntrypointAuthorityMembership(
				"harpy_v2_runtime", "harpy_forms_record_writer", true, false, true,
			),
			builtEntrypointAuthorityMembership(
				"harpy_migrator", "harpy_owner", true, false, true,
			),
			builtEntrypointAuthorityMembership(
				"harpy_backup", "harpy_backup_reader", true, false, true,
			),
			builtEntrypointAuthorityMembership(
				"harpy_restore", "harpy_owner", true, false, true,
			),
			builtEntrypointAuthorityMembership(
				"harpy_v1_runtime", "harpy_runtime", false, true, false,
			),
		},
		CredentialLeases: []runtimev1.BootstrapCredentialLease{
			builtEntrypointAuthorityLease(
				builtEntrypointClusterLease, "harpy_cluster_admin", issuedAt, expiresAt,
			),
			builtEntrypointAuthorityLease(
				builtEntrypointExtensionLease, "harpy_extension_admin", issuedAt, expiresAt,
			),
		},
		RolesManifestSHA256:              catalog.roles.digest,
		PrivilegesManifestSHA256:         catalog.privileges.digest,
		OwnershipManifestSHA256:          catalog.ownership.digest,
		BootstrapOperationDigest:         catalog.bootstrapOperation.digest,
		CredentialCleanupOperationDigest: catalog.cleanupOperation.digest,
	}
}

func writeBuiltEntrypointAuthority(
	t *testing.T,
	directory string,
	now time.Time,
	catalog builtEntrypointArtifactCatalog,
	signer builtEntrypointSigner,
) (builtEntrypointArtifact, builtEntrypointArtifact) {
	t.Helper()
	manifest := builtEntrypointAuthorityManifest(now, catalog)
	if err := runtimev1.ValidateBootstrapAuthorityManifest(manifest); err != nil {
		t.Fatalf("validate built entrypoint authority: %v", err)
	}
	document := writeBuiltEntrypointArtifact(t, directory, "authority.json", manifest)
	signature := runtimev1.BootstrapAuthoritySignature{
		ContractVersion: "bootstrap-authority-signature.v1",
		ManifestSHA256:  document.digest,
		Signature: builtEntrypointAttestationSignature(
			t, signer.signer, document.document, now.Add(-time.Minute),
		),
	}
	signed := writeBuiltEntrypointArtifact(t, directory, "authority.sig", signature)
	return document, signed
}
