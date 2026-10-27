package attestation

import (
	"strings"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func validRuntimeBootstrapReceipt() runtimev1.BootstrapReceipt {
	digest := strings.Repeat("a", 64)
	return runtimev1.BootstrapReceipt{
		ContractVersion: "postgres-bootstrap-receipt.v1", Kind: "cluster_bootstrap",
		Payload: runtimev1.BootstrapReceiptPayload{
			RunID:         "11111111-1111-4111-8111-111111111111",
			ReleaseID:     "44444444-4444-4444-8444-444444444444",
			EnvironmentID: "88888888-8888-4888-8888-888888888888",
			IssuedAt:      "2026-07-13T00:01:00Z", ReleaseManifestSHA256: digest,
			BootstrapAuthorityManifestSHA256: digest,
			CredentialLeaseID:                "55555555-5555-4555-8555-555555555555",
			PostgresImageDigest:              runtimev1.PostgresImageDigest,
			OwnershipManifestSHA256:          digest, DatabaseName: "harpy",
			DatabaseSystemIdentifier: "123456789", DatabaseOID: 16384,
			ServerVersionNumber: runtimev1.PostgresServerVersionNumber,
			DatabaseOrigin:      "fresh", DatabaseOwner: "harpy_owner",
			DatabaseEncoding: "UTF8", DatabaseLocaleProvider: "libc",
			DatabaseCollation: "en_US.utf8", DatabaseCType: "en_US.utf8",
			DataChecksums: true, SessionUser: "harpy_cluster_admin",
			CurrentUser: "harpy_cluster_admin", OperationDigest: digest,
			Result: "succeeded", RolesManifestSHA256: digest,
			PrivilegesManifestSHA256: digest,
		},
	}
}

func validRuntimeExtensionReceipt() runtimev1.ExtensionReceipt {
	digest := strings.Repeat("a", 64)
	return runtimev1.ExtensionReceipt{
		ContractVersion: "postgres-extension-receipt.v1", Kind: "extension_admin",
		Payload: runtimev1.ExtensionReceiptPayload{
			RunID:     "22222222-2222-4222-8222-222222222222",
			ReleaseID: "44444444-4444-4444-8444-444444444444",
			IssuedAt:  "2026-07-13T00:02:00Z", ReleaseManifestSHA256: digest,
			BootstrapAuthorityManifestSHA256: digest,
			CredentialLeaseID:                "66666666-6666-4666-8666-666666666666",
			PostgresImageDigest:              runtimev1.PostgresImageDigest,
			OwnershipManifestSHA256:          digest, DatabaseName: "harpy",
			DatabaseSystemIdentifier: "123456789", DatabaseOID: 16384,
			ServerVersionNumber: runtimev1.PostgresServerVersionNumber,
			DatabaseOrigin:      "fresh", DatabaseOwner: "harpy_owner",
			DatabaseEncoding: "UTF8", DatabaseLocaleProvider: "libc",
			DatabaseCollation: "en_US.utf8", DatabaseCType: "en_US.utf8",
			DataChecksums: true, SessionUser: "harpy_extension_admin",
			CurrentUser: "harpy_extension_admin", OperationDigest: digest,
			Result: "succeeded", ExtensionManifestSHA256: digest,
			Extensions: []runtimev1.ExtensionRecord{{
				Name: "vector", Version: "0.8.3", Schema: "platform_ext",
				Owner: "harpy_extension_admin", ControlFileSHA256: digest,
			}},
		},
	}
}

func validRuntimeCleanupReceipt() runtimev1.CredentialCleanupReceipt {
	digest := strings.Repeat("a", 64)
	return runtimev1.CredentialCleanupReceipt{
		ContractVersion: "credential-cleanup-receipt.v1", Kind: "credential_cleanup",
		Payload: runtimev1.CredentialCleanupReceiptPayload{
			RunID:     "33333333-3333-4333-8333-333333333333",
			ReleaseID: "44444444-4444-4444-8444-444444444444",
			IssuedAt:  "2026-07-13T00:04:00Z", ReleaseManifestSHA256: digest,
			BootstrapAuthorityManifestSHA256: digest, CleanupOperationDigest: digest,
			DatabaseName: "harpy", DatabaseOrigin: "fresh",
			DatabaseSystemIdentifier: "123456789", DatabaseOID: 16384,
			DatabaseOwner: "harpy_owner", DatabaseEncoding: "UTF8",
			DatabaseLocaleProvider: "libc", DatabaseCollation: "en_US.utf8",
			DatabaseCType: "en_US.utf8", DataChecksums: true,
			Principals: []runtimev1.CleanedPrincipal{
				cleanedRuntimePrincipal("harpy_cluster_admin", "55555555-5555-4555-8555-555555555555"),
				cleanedRuntimePrincipal("harpy_extension_admin", "66666666-6666-4666-8666-666666666666"),
			},
			Result: "succeeded",
		},
	}
}

func cleanedRuntimePrincipal(name, leaseID string) runtimev1.CleanedPrincipal {
	return runtimev1.CleanedPrincipal{
		Name: name, LeaseID: leaseID, CleanedAt: "2026-07-13T00:03:00Z",
		CanLogin: false, PasswordState: nil, ActiveSessions: 0, MembershipCount: 0,
	}
}
