package attestation

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

var runtimeTestNow = time.Date(2026, 7, 13, 0, 5, 0, 0, time.UTC)

func runtimeTestRootAndSigner(t *testing.T) (TrustRoot, Signer) {
	t.Helper()
	_, _, root := releaseTestDocuments(t)
	seed := make([]byte, 32)
	for index := range seed {
		seed[index] = byte(index + 1)
	}
	signerDocument := mustRuntimeJSON(t, map[string]any{
		"contract_version": "attestation-signing-key.v1",
		"key_id":           "release-1", "algorithm": "Ed25519",
		"seed": base64.RawURLEncoding.EncodeToString(seed),
	})
	signer, err := ParseSigner(signerDocument, root)
	if err != nil {
		t.Fatalf("parse runtime test signer: %v", err)
	}
	return root, signer
}

func validRuntimeAuthority() runtimev1.BootstrapAuthorityManifest {
	digest := strings.Repeat("a", 64)
	template := "template0"
	return runtimev1.BootstrapAuthorityManifest{
		ContractVersion: "bootstrap-authority-manifest.v1",
		ManifestID:      "77777777-7777-4777-8777-777777777777",
		EnvironmentID:   "88888888-8888-4888-8888-888888888888",
		ReleaseID:       "44444444-4444-4444-8444-444444444444",
		IssuedAt:        "2026-07-13T00:00:00Z", ExpiresAt: "2026-07-13T00:15:00Z",
		DatabaseOrigin: "fresh",
		Database: runtimev1.BootstrapAuthorityDatabase{
			Name: "harpy", CreateDatabase: true, SourceOwner: nil,
			Owner: "harpy_owner", ExtensionOwner: "harpy_extension_admin",
			ExtensionOwnerMode:             "fresh_leased_admin",
			ExtensionOwnerApplicationLogin: false,
			OwnershipExecutor:              "cluster_bootstrap",
			OwnershipAttestation:           "bootstrap_live_catalog",
			Encoding:                       "UTF8", LocaleProvider: "libc", Collation: "en_US.utf8",
			CType: "en_US.utf8", DataChecksums: true, Template: &template,
			PublicConnect: false, PublicCreate: false, PublicTemp: false,
		},
		Connections: runtimev1.BootstrapAuthorityConnections{
			EndpointKind: "direct_postgres", MaintenanceDatabase: "postgres",
			TargetDatabase: "harpy", SameServerRequired: true,
		},
		CreateRoleSelfGrant: "", Roles: validRuntimeAuthorityRoles(),
		Memberships: []runtimev1.BootstrapAuthorityMembership{
			runtimeMembership("harpy_v2_runtime", "harpy_runtime", true, false, true),
			runtimeMembership("harpy_migrator", "harpy_owner", true, false, true),
			runtimeMembership(
				"harpy_backup", "harpy_backup_reader", true, false, true,
			),
			runtimeMembership("harpy_restore", "harpy_owner", true, false, true),
			runtimeMembership("harpy_v1_runtime", "harpy_runtime", false, true, false),
		},
		CredentialLeases: []runtimev1.BootstrapCredentialLease{
			runtimeLease("55555555-5555-4555-8555-555555555555", "harpy_cluster_admin"),
			runtimeLease("66666666-6666-4666-8666-666666666666", "harpy_extension_admin"),
		},
		RolesManifestSHA256: digest, PrivilegesManifestSHA256: digest,
		OwnershipManifestSHA256: digest, BootstrapOperationDigest: digest,
		CredentialCleanupOperationDigest: digest,
	}
}

func validRuntimeAuthorityRoles() []runtimev1.BootstrapAuthorityRole {
	return []runtimev1.BootstrapAuthorityRole{
		runtimeRole("harpy_owner", true, "permanent", false, false, false, false, -1),
		runtimeRole("harpy_cutover_authority_owner", true, "externally_provisioned", false, false, false, false, -1),
		runtimeRole("harpy_runtime", true, "permanent", false, false, false, false, -1),
		runtimeRole("harpy_v2_runtime", true, "permanent", true, false, false, false, -1),
		runtimeRole("harpy_migrator", true, "permanent", true, false, false, false, -1),
		runtimeRole("harpy_v1_runtime", false, "coexistence", false, false, false, false, -1),
		runtimeRole("harpy_cluster_admin", true, "leased", false, false, true, true, 1),
		runtimeRole("harpy_extension_admin", true, "leased", false, true, false, false, 1),
		runtimeRole("harpy_legacy_extension_owner", false, "adopted_only", false, false, false, false, -1),
		runtimeRole("harpy_adopter", true, "cutover_leased", false, false, false, false, 0),
		runtimeRole("harpy_cutover_decoder", true, "cutover_leased", false, false, false, false, 0),
		runtimeRole("harpy_cutover_finalizer", true, "cutover_finalization_leased", false, false, false, false, 0),
		runtimeRole("harpy_backup_reader", true, "permanent", false, false, false, false, -1),
		runtimeRole("harpy_backup", true, "permanent", true, false, false, false, -1),
		runtimeRole("harpy_restore", true, "permanent", true, false, false, false, -1),
		runtimeRole("harpy_reactivation_observer", true, "permanent", true, false, false, false, -1),
	}
}

func runtimeRole(
	name string, enabled bool, mode string, login, superuser, createDB, createRole bool, limit int,
) runtimev1.BootstrapAuthorityRole {
	return runtimev1.BootstrapAuthorityRole{
		Name: name, Enabled: enabled, CredentialMode: mode, Login: login,
		Inherit: false, Superuser: superuser, CreateDB: createDB, CreateRole: createRole,
		Replication: false, BypassRLS: false, ConnectionLimit: limit,
	}
}

func runtimeMembership(
	member, role string, enabled, inherit, set bool,
) runtimev1.BootstrapAuthorityMembership {
	return runtimev1.BootstrapAuthorityMembership{
		Member: member, Role: role, Enabled: enabled, Admin: false,
		Inherit: inherit, Set: set, Grantor: "harpy_cluster_admin",
	}
}

func runtimeLease(id, principal string) runtimev1.BootstrapCredentialLease {
	return runtimev1.BootstrapCredentialLease{
		LeaseID: id, Principal: principal, NotBefore: "2026-07-13T00:00:00Z",
		ExpiresAt: "2026-07-13T00:15:00Z", MaxDurationSeconds: 900,
		MaxConnections: 1, CleanupPolicy: "disable_login_clear_password_terminate_sessions",
	}
}

func mustRuntimeJSON(t *testing.T, value any) []byte {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal runtime test document: %v", err)
	}
	return document
}
