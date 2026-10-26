package attestation

import runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"

type VerifiedBootstrapAuthority struct {
	manifest       runtimev1.BootstrapAuthorityManifest
	manifestDigest string
	verified       bool
}

func (authority VerifiedBootstrapAuthority) Value() runtimev1.BootstrapAuthorityManifest {
	manifest := authority.manifest
	manifest.Roles = append([]runtimev1.BootstrapAuthorityRole(nil), manifest.Roles...)
	manifest.Memberships = append(
		[]runtimev1.BootstrapAuthorityMembership(nil), manifest.Memberships...,
	)
	manifest.CredentialLeases = append(
		[]runtimev1.BootstrapCredentialLease(nil), manifest.CredentialLeases...,
	)
	manifest.Database.SourceOwner = cloneStringPointer(manifest.Database.SourceOwner)
	manifest.Database.Template = cloneStringPointer(manifest.Database.Template)
	return manifest
}

func (authority VerifiedBootstrapAuthority) ManifestSHA256() string {
	return authority.manifestDigest
}

type VerifiedBootstrapReceipt struct {
	receipt  runtimev1.BootstrapReceipt
	verified bool
}

func (receipt VerifiedBootstrapReceipt) Value() runtimev1.BootstrapReceipt {
	return receipt.receipt
}

type VerifiedExtensionReceipt struct {
	receipt  runtimev1.ExtensionReceipt
	verified bool
}

func (receipt VerifiedExtensionReceipt) Value() runtimev1.ExtensionReceipt {
	value := receipt.receipt
	value.Payload.Extensions = append(
		[]runtimev1.ExtensionRecord(nil), receipt.receipt.Payload.Extensions...,
	)
	return value
}

type VerifiedCredentialCleanupReceipt struct {
	receipt  runtimev1.CredentialCleanupReceipt
	verified bool
}

func (receipt VerifiedCredentialCleanupReceipt) Value() runtimev1.CredentialCleanupReceipt {
	value := receipt.receipt
	value.Payload.Principals = append(
		[]runtimev1.CleanedPrincipal(nil), receipt.receipt.Payload.Principals...,
	)
	for index := range value.Payload.Principals {
		value.Payload.Principals[index].PasswordState = cloneStringPointer(
			value.Payload.Principals[index].PasswordState,
		)
	}
	return value
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
