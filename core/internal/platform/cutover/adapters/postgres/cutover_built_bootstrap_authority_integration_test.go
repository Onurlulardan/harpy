//go:build integration

package postgres

import (
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func writeBuiltCutoverBootstrapAuthority(
	t *testing.T, directory string, signer builtCutoverSigner, now time.Time,
) (builtCutoverArtifact, builtCutoverArtifact) {
	t.Helper()
	issuedAt := now.Add(-time.Minute).UTC().Truncate(time.Second)
	expiresAt := now.Add(9 * time.Minute).UTC().Truncate(time.Second)
	template := "template0"
	digest := digestMarker('a')
	manifest := runtimev1.BootstrapAuthorityManifest{
		ContractVersion: "bootstrap-authority-manifest.v1",
		ManifestID:      "32000000-0000-4000-8000-000000000001",
		EnvironmentID:   integrationEnvironment, ReleaseID: integrationReleaseID,
		IssuedAt: issuedAt.Format(time.RFC3339), ExpiresAt: expiresAt.Format(time.RFC3339),
		DatabaseOrigin: "fresh",
		Database: runtimev1.BootstrapAuthorityDatabase{
			Name: "harpy_cutover", CreateDatabase: true, Owner: "harpy_owner",
			ExtensionOwner:       "harpy_extension_admin",
			ExtensionOwnerMode:   "fresh_leased_admin",
			OwnershipExecutor:    "cluster_bootstrap",
			OwnershipAttestation: "bootstrap_live_catalog",
			Encoding:             runtimev1.PostgresEncoding, LocaleProvider: runtimev1.PostgresLocaleProvider,
			Collation: runtimev1.PostgresCollation, CType: runtimev1.PostgresCharacterType,
			DataChecksums: runtimev1.PostgresDataChecksums, Template: &template,
		},
		Connections: runtimev1.BootstrapAuthorityConnections{
			EndpointKind: "direct_postgres", MaintenanceDatabase: "postgres",
			TargetDatabase: "harpy_cutover", SameServerRequired: true,
		},
		Roles: builtCutoverAuthorityRoles(),
		Memberships: []runtimev1.BootstrapAuthorityMembership{
			builtCutoverAuthorityMembership("harpy_v2_runtime", "harpy_runtime", true, false, true),
			builtCutoverAuthorityMembership("harpy_migrator", "harpy_owner", true, false, true),
			builtCutoverAuthorityMembership("harpy_backup", "harpy_backup_reader", true, false, true),
			builtCutoverAuthorityMembership("harpy_restore", "harpy_owner", true, false, true),
			builtCutoverAuthorityMembership("harpy_v1_runtime", "harpy_runtime", false, true, false),
		},
		CredentialLeases: []runtimev1.BootstrapCredentialLease{
			builtCutoverAuthorityLease("32000000-0000-4000-8000-000000000002",
				"harpy_cluster_admin", issuedAt, expiresAt),
			builtCutoverAuthorityLease("32000000-0000-4000-8000-000000000003",
				"harpy_extension_admin", issuedAt, expiresAt),
		},
		RolesManifestSHA256: digest, PrivilegesManifestSHA256: digest,
		OwnershipManifestSHA256: digest, BootstrapOperationDigest: digest,
		CredentialCleanupOperationDigest: digest,
	}
	if err := runtimev1.ValidateBootstrapAuthorityManifest(manifest); err != nil {
		t.Fatalf("validate built bootstrap authority: %v", err)
	}
	document := writeBuiltCutoverArtifact(t, directory, "bootstrap-authority.json", manifest)
	signature := signBuiltCutoverDocument(t, signer, document.document, issuedAt)
	signed := writeBuiltCutoverArtifact(t, directory, "bootstrap-authority.sig",
		runtimev1.BootstrapAuthoritySignature{
			ContractVersion: "bootstrap-authority-signature.v1",
			ManifestSHA256:  document.digest,
			Signature: runtimev1.AttestationSignature{
				Algorithm: signature.Algorithm, KeyID: signature.KeyID, Value: signature.Value,
			},
		})
	return document, signed
}

func builtCutoverAuthorityLease(
	id, principal string, issuedAt, expiresAt time.Time,
) runtimev1.BootstrapCredentialLease {
	return runtimev1.BootstrapCredentialLease{
		LeaseID: id, Principal: principal, NotBefore: issuedAt.Format(time.RFC3339),
		ExpiresAt: expiresAt.Format(time.RFC3339), MaxDurationSeconds: 900,
		MaxConnections: 1,
		CleanupPolicy:  "disable_login_clear_password_terminate_sessions",
	}
}
