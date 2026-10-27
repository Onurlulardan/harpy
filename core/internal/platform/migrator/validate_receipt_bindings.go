package migrator

import (
	"fmt"
	"slices"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func validateReceiptBindings(
	release attestation.ReleaseManifest,
	releaseDigest string,
	authority runtimev1.BootstrapAuthorityManifest,
	artifacts attestation.VerifiedPostgresArtifactSet,
	bootstrap runtimev1.BootstrapReceipt,
	extension runtimev1.ExtensionReceipt,
	cleanup runtimev1.CredentialCleanupReceipt,
) error {
	if err := runtimev1.ValidateBootstrapReceipt(bootstrap); err != nil {
		return fmt.Errorf("bootstrap receipt is not verified evidence: %w", err)
	}
	if err := runtimev1.ValidateExtensionReceipt(extension); err != nil {
		return fmt.Errorf("extension receipt is not verified evidence: %w", err)
	}
	if err := runtimev1.ValidateCredentialCleanupReceipt(cleanup); err != nil {
		return fmt.Errorf("cleanup receipt is not verified evidence: %w", err)
	}
	binding := receiptBinding{
		release: release, releaseDigest: releaseDigest,
		authority: authority, artifacts: artifacts,
		bootstrap: bootstrap.Payload, extension: extension.Payload,
		cleanup: cleanup.Payload,
	}
	if err := binding.validateIdentity(); err != nil {
		return err
	}
	if err := binding.validateDigests(); err != nil {
		return err
	}
	return binding.validateExtensions()
}

type receiptBinding struct {
	release       attestation.ReleaseManifest
	releaseDigest string
	authority     runtimev1.BootstrapAuthorityManifest
	artifacts     attestation.VerifiedPostgresArtifactSet
	bootstrap     runtimev1.BootstrapReceiptPayload
	extension     runtimev1.ExtensionReceiptPayload
	cleanup       runtimev1.CredentialCleanupReceiptPayload
}

func (binding receiptBinding) validateIdentity() error {
	release := binding.release
	authority := binding.authority
	bootstrap := binding.bootstrap
	extension := binding.extension
	cleanup := binding.cleanup
	if bootstrap.ReleaseID != release.ReleaseID || extension.ReleaseID != release.ReleaseID ||
		cleanup.ReleaseID != release.ReleaseID || bootstrap.EnvironmentID != authority.EnvironmentID {
		return fmt.Errorf("receipt release or environment identity differs")
	}
	if bootstrap.DatabaseName != release.DatabaseName ||
		extension.DatabaseName != release.DatabaseName || cleanup.DatabaseName != release.DatabaseName ||
		bootstrap.DatabaseOrigin != release.DatabaseOrigin ||
		extension.DatabaseOrigin != release.DatabaseOrigin || cleanup.DatabaseOrigin != release.DatabaseOrigin {
		return fmt.Errorf("receipt database identity differs")
	}
	if bootstrap.DatabaseSystemIdentifier != extension.DatabaseSystemIdentifier ||
		bootstrap.DatabaseSystemIdentifier != cleanup.DatabaseSystemIdentifier ||
		bootstrap.DatabaseOID != extension.DatabaseOID || bootstrap.DatabaseOID != cleanup.DatabaseOID {
		return fmt.Errorf("receipt physical database identity differs")
	}
	if !sameReceiptDatabaseShape(release, bootstrap, extension, cleanup) {
		return fmt.Errorf("receipt database shape differs from release")
	}
	bootstrapAt, _ := time.Parse(time.RFC3339Nano, bootstrap.IssuedAt)
	extensionAt, _ := time.Parse(time.RFC3339Nano, extension.IssuedAt)
	cleanupAt, _ := time.Parse(time.RFC3339Nano, cleanup.IssuedAt)
	if extensionAt.Before(bootstrapAt) || cleanupAt.Before(extensionAt) {
		return fmt.Errorf("receipt evidence order is invalid")
	}
	return nil
}

func sameReceiptDatabaseShape(
	release attestation.ReleaseManifest,
	bootstrap runtimev1.BootstrapReceiptPayload,
	extension runtimev1.ExtensionReceiptPayload,
	cleanup runtimev1.CredentialCleanupReceiptPayload,
) bool {
	return bootstrap.DatabaseOwner == release.DatabaseOwner &&
		extension.DatabaseOwner == release.DatabaseOwner && cleanup.DatabaseOwner == release.DatabaseOwner &&
		bootstrap.DatabaseEncoding == release.DatabaseEncoding &&
		extension.DatabaseEncoding == release.DatabaseEncoding && cleanup.DatabaseEncoding == release.DatabaseEncoding &&
		bootstrap.DatabaseLocaleProvider == release.DatabaseLocaleProvider &&
		extension.DatabaseLocaleProvider == release.DatabaseLocaleProvider && cleanup.DatabaseLocaleProvider == release.DatabaseLocaleProvider &&
		bootstrap.DatabaseCollation == release.DatabaseCollation &&
		extension.DatabaseCollation == release.DatabaseCollation && cleanup.DatabaseCollation == release.DatabaseCollation &&
		bootstrap.DatabaseCType == release.DatabaseCType &&
		extension.DatabaseCType == release.DatabaseCType && cleanup.DatabaseCType == release.DatabaseCType &&
		bootstrap.DataChecksums == release.DataChecksums &&
		extension.DataChecksums == release.DataChecksums && cleanup.DataChecksums == release.DataChecksums
}

func (binding receiptBinding) validateDigests() error {
	release := binding.release
	authorityDigest := release.BootstrapAuthoritySHA256
	for _, actual := range []string{
		binding.bootstrap.ReleaseManifestSHA256,
		binding.extension.ReleaseManifestSHA256,
		binding.cleanup.ReleaseManifestSHA256,
	} {
		if actual != binding.releaseDigest {
			return fmt.Errorf("receipt release digest differs")
		}
	}
	if binding.bootstrap.BootstrapAuthorityManifestSHA256 != authorityDigest ||
		binding.extension.BootstrapAuthorityManifestSHA256 != authorityDigest ||
		binding.cleanup.BootstrapAuthorityManifestSHA256 != authorityDigest {
		return fmt.Errorf("receipt authority digest differs")
	}
	if binding.bootstrap.OperationDigest != release.BootstrapOperationDigest ||
		binding.extension.OperationDigest != release.ExtensionOperationDigest ||
		binding.cleanup.CleanupOperationDigest != release.CredentialCleanupDigest {
		return fmt.Errorf("receipt operation digest differs")
	}
	if binding.bootstrap.RolesManifestSHA256 != release.RolesManifestSHA256 ||
		binding.bootstrap.PrivilegesManifestSHA256 != release.PrivilegesManifestSHA256 ||
		binding.extension.ExtensionManifestSHA256 != release.ExtensionManifestSHA256 ||
		binding.bootstrap.OwnershipManifestSHA256 != release.OwnershipManifestSHA256 ||
		binding.extension.OwnershipManifestSHA256 != release.OwnershipManifestSHA256 {
		return fmt.Errorf("receipt artifact digest differs")
	}
	return nil
}

func (binding receiptBinding) validateExtensions() error {
	expected := binding.artifacts.Artifacts().Extension.Value().Extensions
	actual := binding.extension.Extensions
	if len(expected) != len(actual) {
		return fmt.Errorf("extension receipt inventory differs from manifest")
	}
	for index, extension := range expected {
		record := actual[index]
		if extension.Name != record.Name ||
			extension.Version != record.Version || extension.Schema != record.Schema ||
			extension.Owner != record.Owner || extension.ControlFileSHA256 != record.ControlFileSHA256 ||
			!slices.IsSorted(extension.Capabilities) {
			return fmt.Errorf("extension receipt record %d differs from manifest", index)
		}
	}
	return nil
}
