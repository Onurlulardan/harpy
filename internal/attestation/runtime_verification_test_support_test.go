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
			runtimeMembership("harpy_goose", "harpy_owner"),
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
		runtimeRole("harpy_owner", false, false, -1),
		runtimeRole("harpy_goose", true, false, 4),
		runtimeRole("harpy_api", true, true, 20),
		runtimeRole("harpy_worker", true, true, 20),
	}
}

func runtimeRole(name string, login, inherit bool, limit int) runtimev1.BootstrapAuthorityRole {
	return runtimev1.BootstrapAuthorityRole{
		Name: name, Enabled: true, CredentialMode: "permanent", Login: login,
		Inherit: inherit, Superuser: false, CreateDB: false, CreateRole: false,
		Replication: false, BypassRLS: false, ConnectionLimit: limit,
	}
}

func runtimeMembership(member, role string) runtimev1.BootstrapAuthorityMembership {
	return runtimev1.BootstrapAuthorityMembership{
		Member: member, Role: role, Enabled: true, Admin: false,
		Inherit: false, Set: true, Grantor: "harpy_owner",
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
