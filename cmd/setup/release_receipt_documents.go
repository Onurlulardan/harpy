package main

import (
	"os"
	"strings"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func buildRuntimeReceipts(
	keys releaseKeys, manifests postgresManifests, operations postgresOperations,
	authority releaseDocument, release releaseDocument,
) (bootstrap, extension, cleanup []byte, err error) {
	bootstrap, err = signedBootstrapReceipt(keys, manifests, operations, authority, release)
	if err != nil {
		return nil, nil, nil, err
	}
	extension, err = signedExtensionReceipt(keys, manifests, operations, authority, release)
	if err != nil {
		return nil, nil, nil, err
	}
	cleanup, err = signedCleanupReceipt(keys, operations, authority, release)
	if err != nil {
		return nil, nil, nil, err
	}
	return bootstrap, extension, cleanup, nil
}

func signedBootstrapReceipt(
	keys releaseKeys, manifests postgresManifests, operations postgresOperations,
	authority, release releaseDocument,
) ([]byte, error) {
	payload := runtimev1.BootstrapReceiptPayload{
		RunID: localBootstrapRun, ReleaseID: localReleaseID, EnvironmentID: localEnvironmentID,
		IssuedAt: keys.clock.stamp(keys.clock.issued), ReleaseManifestSHA256: release.digest,
		BootstrapAuthorityManifestSHA256: authority.digest, CredentialLeaseID: localClusterLease,
		PostgresImageDigest: runtimev1.PostgresImageDigest, OwnershipManifestSHA256: manifests.ownership.digest,
		DatabaseName: localDatabase, DatabaseSystemIdentifier: databaseSystemID(), DatabaseOID: localDatabaseOID,
		ServerVersionNumber: runtimev1.PostgresServerVersionNumber, DatabaseOrigin: "fresh",
		DatabaseOwner: runtimev1.DatabaseOwnerRole, DatabaseEncoding: runtimev1.PostgresEncoding,
		DatabaseLocaleProvider: runtimev1.PostgresLocaleProvider, DatabaseCollation: runtimev1.PostgresCollation,
		DatabaseCType: runtimev1.PostgresCharacterType, DataChecksums: runtimev1.PostgresDataChecksums,
		SessionUser: "harpy_cluster_admin", CurrentUser: "harpy_cluster_admin",
		OperationDigest: operations.bootstrap.digest, Result: "succeeded",
		RolesManifestSHA256: manifests.roles.digest, PrivilegesManifestSHA256: manifests.privileges.digest,
	}
	return finishReceipt(keys, "postgres-bootstrap-receipt.v1", "cluster_bootstrap", payload, func(body []byte) error {
		_, err := attestation.VerifyBootstrapReceipt(body, keys.root, keys.clock.now)
		return err
	})
}

func signedExtensionReceipt(
	keys releaseKeys, manifests postgresManifests, operations postgresOperations,
	authority, release releaseDocument,
) ([]byte, error) {
	records, err := extensionReceiptRecords(manifests.extension)
	if err != nil {
		return nil, err
	}
	payload := runtimev1.ExtensionReceiptPayload{
		RunID: localExtensionRun, ReleaseID: localReleaseID, IssuedAt: keys.clock.stamp(keys.clock.issued),
		ReleaseManifestSHA256: release.digest, BootstrapAuthorityManifestSHA256: authority.digest,
		CredentialLeaseID: localExtensionLease, PostgresImageDigest: runtimev1.PostgresImageDigest,
		OwnershipManifestSHA256: manifests.ownership.digest, DatabaseName: localDatabase,
		DatabaseSystemIdentifier: databaseSystemID(), DatabaseOID: localDatabaseOID,
		ServerVersionNumber: runtimev1.PostgresServerVersionNumber, DatabaseOrigin: "fresh",
		DatabaseOwner: runtimev1.DatabaseOwnerRole, DatabaseEncoding: runtimev1.PostgresEncoding,
		DatabaseLocaleProvider: runtimev1.PostgresLocaleProvider, DatabaseCollation: runtimev1.PostgresCollation,
		DatabaseCType: runtimev1.PostgresCharacterType, DataChecksums: runtimev1.PostgresDataChecksums,
		SessionUser: "harpy_extension_admin", CurrentUser: "harpy_extension_admin",
		OperationDigest: operations.extension.digest, Result: "succeeded",
		ExtensionManifestSHA256: manifests.extension.digest, Extensions: records,
	}
	return finishReceipt(keys, "postgres-extension-receipt.v1", "extension_admin", payload, func(body []byte) error {
		_, err := attestation.VerifyExtensionReceipt(body, keys.root, keys.clock.now)
		return err
	})
}

func signedCleanupReceipt(
	keys releaseKeys, operations postgresOperations, authority, release releaseDocument,
) ([]byte, error) {
	payload := runtimev1.CredentialCleanupReceiptPayload{
		RunID: localCleanupRun, ReleaseID: localReleaseID, IssuedAt: keys.clock.stamp(keys.clock.issued),
		ReleaseManifestSHA256: release.digest, BootstrapAuthorityManifestSHA256: authority.digest,
		CleanupOperationDigest: operations.cleanup.digest, DatabaseName: localDatabase,
		DatabaseOrigin: "fresh", DatabaseSystemIdentifier: databaseSystemID(), DatabaseOID: localDatabaseOID,
		DatabaseOwner: runtimev1.DatabaseOwnerRole, DatabaseEncoding: runtimev1.PostgresEncoding,
		DatabaseLocaleProvider: runtimev1.PostgresLocaleProvider, DatabaseCollation: runtimev1.PostgresCollation,
		DatabaseCType: runtimev1.PostgresCharacterType, DataChecksums: runtimev1.PostgresDataChecksums,
		Result: "succeeded",
		Principals: []runtimev1.CleanedPrincipal{
			cleanedPrincipal(keys, "harpy_cluster_admin", localClusterLease),
			cleanedPrincipal(keys, "harpy_extension_admin", localExtensionLease),
		},
	}
	return finishReceipt(keys, "credential-cleanup-receipt.v1", "credential_cleanup", payload, func(body []byte) error {
		_, err := attestation.VerifyCredentialCleanupReceipt(body, keys.root, keys.clock.now)
		return err
	})
}

func databaseSystemID() string {
	value := strings.TrimSpace(os.Getenv("HARPY_DATABASE_SYSTEM_IDENTIFIER"))
	if value == "" {
		return localSystemID
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return localSystemID
		}
	}
	return value
}

func extensionReceiptRecords(manifest releaseDocument) ([]runtimev1.ExtensionRecord, error) {
	var decoded postgresv1.ExtensionManifest
	if err := attestation.DecodeStrict(manifest.body, &decoded); err != nil {
		return nil, err
	}
	records := make([]runtimev1.ExtensionRecord, 0, len(decoded.Extensions))
	for _, extension := range decoded.Extensions {
		records = append(records, runtimev1.ExtensionRecord{
			Name: extension.Name, Version: extension.Version, Schema: extension.Schema,
			Owner: extension.Owner, ControlFileSHA256: extension.ControlFileSHA256,
		})
	}
	return records, nil
}

func cleanedPrincipal(keys releaseKeys, name, lease string) runtimev1.CleanedPrincipal {
	return runtimev1.CleanedPrincipal{
		Name: name, LeaseID: lease, CleanedAt: keys.clock.stamp(keys.clock.cleaned),
	}
}

func finishReceipt[T any](
	keys releaseKeys, contract, kind string, payload T, verify func([]byte) error,
) ([]byte, error) {
	signature, err := signPayload(keys, contract, kind, payload)
	if err != nil {
		return nil, err
	}
	type document struct {
		ContractVersion string                         `json:"contract_version"`
		Kind            string                         `json:"kind"`
		Payload         T                              `json:"payload"`
		Signature       runtimev1.AttestationSignature `json:"signature"`
	}
	body, err := attestation.MarshalCanonical(document{
		ContractVersion: contract, Kind: kind, Payload: payload, Signature: signature,
	})
	if err != nil {
		return nil, err
	}
	if err = verify(body); err != nil {
		return nil, err
	}
	return body, nil
}
