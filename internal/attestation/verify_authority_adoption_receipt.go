package attestation

import (
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func VerifyAuthorityAdoptionReceipt(
	document, signatureDocument []byte,
	operation VerifiedAuthorityAdoptionOperation,
	release VerifiedRelease,
	authority VerifiedBootstrapAuthority,
	root TrustRoot,
	now time.Time,
) (VerifiedAuthorityAdoptionReceipt, error) {
	if !operation.verified || !release.verified || !authority.verified {
		return VerifiedAuthorityAdoptionReceipt{}, fmt.Errorf("authority receipt prerequisites were not verified")
	}
	var receipt cutoverv1.AuthorityAdoptionReceipt
	if err := DecodeStrict(document, &receipt); err != nil {
		return VerifiedAuthorityAdoptionReceipt{}, fmt.Errorf("decode authority adoption receipt: %w", err)
	}
	if err := receipt.Validate(); err != nil {
		return VerifiedAuthorityAdoptionReceipt{}, err
	}
	adoptedAt, _ := time.Parse(time.RFC3339Nano, receipt.AdoptedAt)
	if adoptedAt.After(now.Add(maximumReleaseClockSkew)) {
		return VerifiedAuthorityAdoptionReceipt{}, fmt.Errorf("authority adoption receipt is from the future")
	}
	digest, err := verifyAuthorityAdoptionSignature(
		document, signatureDocument,
		"cutover-authority-adoption-receipt-signature.v1", adoptedAt, root,
	)
	if err != nil {
		return VerifiedAuthorityAdoptionReceipt{}, err
	}
	if err := bindAuthorityReceipt(receipt, operation, release, authority); err != nil {
		return VerifiedAuthorityAdoptionReceipt{}, err
	}
	canonical, err := Canonicalize(document)
	if err != nil {
		return VerifiedAuthorityAdoptionReceipt{}, err
	}
	return newVerifiedAuthorityAdoptionReceipt(receipt, digest, canonical), nil
}

func bindAuthorityReceipt(
	receipt cutoverv1.AuthorityAdoptionReceipt,
	verifiedOperation VerifiedAuthorityAdoptionOperation,
	release VerifiedRelease,
	authority VerifiedBootstrapAuthority,
) error {
	operation := verifiedOperation.Value()
	if !authorityReceiptWithinOperation(receipt.AdoptedAt, operation.IssuedAt, operation.ExpiresAt) {
		return fmt.Errorf("authority adoption receipt is outside operation window")
	}
	if receipt.OperationID != operation.OperationID || receipt.ReleaseID != operation.ReleaseID ||
		receipt.EnvironmentID != operation.EnvironmentID || receipt.Target != operation.Target ||
		receipt.OperationSHA256 != verifiedOperation.ManifestSHA256() ||
		receipt.MigrationManifestSHA256 != operation.MigrationManifestSHA256 ||
		receipt.ReleaseManifestSHA256 != release.ManifestSHA256() ||
		receipt.BootstrapAuthoritySHA256 != authority.ManifestSHA256() ||
		receipt.Mode != operation.Mode ||
		receipt.PreviousCatalogGeneration != operation.PreviousCatalogGeneration ||
		receipt.CatalogGeneration != operation.CatalogGeneration ||
		receipt.PreCatalogContractSHA256 != operation.PreCatalogContractSHA256 ||
		receipt.PostCatalogContractSHA256 != operation.PostCatalogContractSHA256 ||
		!sameOptionalDigest(receipt.UpgradeArtifactSHA256, operation.UpgradeArtifactSHA256) ||
		receipt.ExecutorSessionUser != operation.ExecutorSessionUser ||
		receipt.AuthorityOwnerRole != operation.AuthorityOwnerRole ||
		receipt.DependencyContractSHA256 != operation.DependencyContractSHA256 ||
		!authorityReceiptObjectsMatch(receipt.Objects, operation.Objects) ||
		!authorityReceiptDependenciesMatch(receipt.Dependencies, operation.Dependencies) {
		return fmt.Errorf("authority adoption receipt binding differs")
	}
	if operation.Mode == "initial_adoption" {
		if receipt.PreviousAuthorityReceiptSHA256 != nil ||
			receipt.PreviousCatalogCheckpoint != nil {
			return fmt.Errorf("initial authority adoption has prior receipt")
		}
		return cutoverv1.ValidateAuthorityQuiescenceBinding(
			operation.QuiescencePolicy, receipt.QuiescenceEvidence,
		)
	}
	if err := bindAuthorityReceiptCheckpoint(receipt, operation); err != nil {
		return err
	}
	return cutoverv1.ValidateAuthorityQuiescenceBinding(
		operation.QuiescencePolicy, receipt.QuiescenceEvidence,
	)
}

func authorityReceiptDependenciesMatch(
	receipts []cutoverv1.AuthorityCatalogDependencyReceipt,
	operations []cutoverv1.AuthorityCatalogDependency,
) bool {
	if len(receipts) != len(operations) {
		return false
	}
	for index, receipt := range receipts {
		if receipt.ObjectOID == 0 || receipt.AuthorityCatalogDependency != operations[index] {
			return false
		}
	}
	return true
}

func authorityReceiptWithinOperation(adoptedAt, issuedAt, expiresAt string) bool {
	adopted, adoptedErr := time.Parse(time.RFC3339Nano, adoptedAt)
	issued, issuedErr := time.Parse(time.RFC3339Nano, issuedAt)
	expires, expiresErr := time.Parse(time.RFC3339Nano, expiresAt)
	return adoptedErr == nil && issuedErr == nil && expiresErr == nil &&
		!adopted.Before(issued) && adopted.Before(expires)
}

func sameOptionalDigest(left, right *string) bool {
	return (left == nil && right == nil) ||
		(left != nil && right != nil && *left == *right)
}

func authorityReceiptObjectsMatch(
	receipts []cutoverv1.AuthorityAdoptionReceiptObject,
	operations []cutoverv1.AuthorityAdoptionObject,
) bool {
	if len(receipts) != len(operations) {
		return false
	}
	for index, receipt := range receipts {
		operation := operations[index]
		if receipt.Kind != operation.Kind || receipt.SchemaName != operation.SchemaName ||
			receipt.ObjectName != operation.ObjectName ||
			receipt.IdentityArguments != operation.IdentityArguments ||
			receipt.IntroducedGeneration != operation.IntroducedGeneration ||
			receipt.Owner != operation.AdoptedOwner ||
			receipt.DefinitionSHA256 != operation.AdoptedDefinitionSHA256 ||
			receipt.ACLSHA256 != operation.AdoptedACLSHA256 {
			return false
		}
	}
	return true
}
