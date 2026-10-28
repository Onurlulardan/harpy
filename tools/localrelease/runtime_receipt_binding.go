package main

import (
	"fmt"
	"slices"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func verifyRuntimeReceiptBindings(
	release attestation.ReleaseManifest, releaseDigest string,
	authority runtimev1.BootstrapAuthorityManifest, catalog releaseCatalog,
	bootstrap runtimev1.BootstrapReceipt, extension runtimev1.ExtensionReceipt,
	cleanup runtimev1.CredentialCleanupReceipt,
) error {
	b, e, c := bootstrap.Payload, extension.Payload, cleanup.Payload
	if b.ReleaseID != release.ReleaseID || e.ReleaseID != release.ReleaseID ||
		c.ReleaseID != release.ReleaseID || b.EnvironmentID != authority.EnvironmentID {
		return fmt.Errorf("runtime receipt release identity differs")
	}
	if b.DatabaseName != release.DatabaseName || e.DatabaseName != release.DatabaseName ||
		c.DatabaseName != release.DatabaseName || b.DatabaseOrigin != release.DatabaseOrigin ||
		e.DatabaseOrigin != release.DatabaseOrigin || c.DatabaseOrigin != release.DatabaseOrigin ||
		b.DatabaseSystemIdentifier != e.DatabaseSystemIdentifier ||
		b.DatabaseSystemIdentifier != c.DatabaseSystemIdentifier ||
		b.DatabaseOID != e.DatabaseOID || b.DatabaseOID != c.DatabaseOID {
		return fmt.Errorf("runtime receipt database identity differs")
	}
	if !runtimeReceiptShapeMatches(release, b, e, c) {
		return fmt.Errorf("runtime receipt database shape differs")
	}
	bootstrapAt, _ := time.Parse(time.RFC3339Nano, b.IssuedAt)
	extensionAt, _ := time.Parse(time.RFC3339Nano, e.IssuedAt)
	cleanupAt, _ := time.Parse(time.RFC3339Nano, c.IssuedAt)
	if extensionAt.Before(bootstrapAt) || cleanupAt.Before(extensionAt) {
		return fmt.Errorf("runtime receipt order differs")
	}
	for _, actual := range []string{
		b.ReleaseManifestSHA256, e.ReleaseManifestSHA256, c.ReleaseManifestSHA256,
	} {
		if actual != releaseDigest {
			return fmt.Errorf("runtime receipt release digest differs")
		}
	}
	if b.BootstrapAuthorityManifestSHA256 != release.BootstrapAuthoritySHA256 ||
		e.BootstrapAuthorityManifestSHA256 != release.BootstrapAuthoritySHA256 ||
		c.BootstrapAuthorityManifestSHA256 != release.BootstrapAuthoritySHA256 ||
		b.OperationDigest != catalog.bootstrapOperation.digest ||
		e.OperationDigest != catalog.extensionOperation.digest ||
		c.CleanupOperationDigest != catalog.cleanupOperation.digest ||
		b.RolesManifestSHA256 != catalog.roles.digest ||
		b.PrivilegesManifestSHA256 != catalog.privileges.digest ||
		b.OwnershipManifestSHA256 != catalog.ownership.digest ||
		e.OwnershipManifestSHA256 != catalog.ownership.digest ||
		e.ExtensionManifestSHA256 != catalog.extension.digest {
		return fmt.Errorf("runtime receipt artifact binding differs")
	}
	expected := extensionManifest().Extensions
	if len(expected) != len(e.Extensions) {
		return fmt.Errorf("runtime extension receipt inventory differs")
	}
	for index, item := range expected {
		record := e.Extensions[index]
		if item.Name != record.Name || item.Version != record.Version ||
			item.Schema != record.Schema || item.Owner != record.Owner ||
			item.ControlFileSHA256 != record.ControlFileSHA256 ||
			!slices.IsSorted(item.Capabilities) {
			return fmt.Errorf("runtime extension receipt record %d differs", index)
		}
	}
	return nil
}

func runtimeReceiptShapeMatches(
	release attestation.ReleaseManifest,
	b runtimev1.BootstrapReceiptPayload, e runtimev1.ExtensionReceiptPayload,
	c runtimev1.CredentialCleanupReceiptPayload,
) bool {
	return b.DatabaseOwner == release.DatabaseOwner && e.DatabaseOwner == release.DatabaseOwner &&
		c.DatabaseOwner == release.DatabaseOwner && b.DatabaseEncoding == release.DatabaseEncoding &&
		e.DatabaseEncoding == release.DatabaseEncoding && c.DatabaseEncoding == release.DatabaseEncoding &&
		b.DatabaseLocaleProvider == release.DatabaseLocaleProvider &&
		e.DatabaseLocaleProvider == release.DatabaseLocaleProvider &&
		c.DatabaseLocaleProvider == release.DatabaseLocaleProvider &&
		b.DatabaseCollation == release.DatabaseCollation &&
		e.DatabaseCollation == release.DatabaseCollation && c.DatabaseCollation == release.DatabaseCollation &&
		b.DatabaseCType == release.DatabaseCType && e.DatabaseCType == release.DatabaseCType &&
		c.DatabaseCType == release.DatabaseCType && b.DataChecksums == release.DataChecksums &&
		e.DataChecksums == release.DataChecksums && c.DataChecksums == release.DataChecksums
}
