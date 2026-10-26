package cutoverv1

import "strings"

func (value AuthorityAdoptionReceipt) Validate() error {
	if value.ContractVersion != "cutover-authority-adoption-receipt.v1" ||
		!validUUID(value.ReceiptID) || !validUUID(value.OperationID) ||
		!validUUID(value.ReleaseID) || !validUUID(value.EnvironmentID) ||
		!validDigest(value.OperationSHA256) || !validDigest(value.MigrationManifestSHA256) ||
		!validDigest(value.ReleaseManifestSHA256) ||
		!validDigest(value.BootstrapAuthoritySHA256) ||
		!validDigest(value.PreCatalogContractSHA256) ||
		!validDigest(value.PostCatalogContractSHA256) ||
		!validDigest(value.CatalogInstanceSHA256) ||
		ValidateDatabaseIdentity(value.Target) != nil {
		return contractError("authority adoption receipt header")
	}
	if _, err := parseUTC(value.AdoptedAt); err != nil {
		return contractError("authority adoption receipt timestamp")
	}
	if value.Result != "succeeded" || !value.ExecutorSuperuser ||
		!sqlNamePattern.MatchString(value.ExecutorSessionUser) ||
		strings.HasPrefix(value.ExecutorSessionUser, "harpy_") ||
		value.PreAdoptionMembershipCount != 0 || value.AuthorityOwnerMembershipCount != 0 ||
		validateAuthorityOwnerRole(value.AuthorityOwnerRole, len(value.Objects)) != nil {
		return contractError("authority adoption receipt executor")
	}
	if value.PostCatalogContractSHA256 != AuthorityReceiptCatalogContractSHA256(value.Objects) ||
		value.CatalogInstanceSHA256 != AuthorityReceiptCatalogInstanceSHA256(value.Objects) {
		return contractError("authority adoption receipt catalog digest")
	}
	if err := validateAuthorityReceiptMode(value); err != nil {
		return err
	}
	if err := validateAuthorityReceiptObjects(value.Objects, value.CatalogGeneration); err != nil {
		return err
	}
	if err := validateAuthorityQuiescenceEvidence(value.QuiescenceEvidence, value.Objects); err != nil {
		return err
	}
	if err := validateAuthorityReceiptDependencies(value); err != nil {
		return err
	}
	return validateAuthorityCatalogCheckpoint(value)
}

func validateAuthorityReceiptMode(value AuthorityAdoptionReceipt) error {
	previousObjects := 0
	for _, object := range value.Objects {
		if object.IntroducedGeneration <= value.PreviousCatalogGeneration {
			previousObjects++
		}
	}
	if value.PreAdoptionOwnedObjectCount != previousObjects {
		return contractError("authority adoption prior object count")
	}
	switch value.Mode {
	case "initial_adoption":
		if value.PreviousCatalogGeneration != 0 || value.CatalogGeneration != 1 ||
			value.PreCatalogContractSHA256 != emptyCatalogSHA256 ||
			value.PreviousAuthorityReceiptSHA256 != nil || value.UpgradeArtifactSHA256 != nil ||
			value.PreviousCatalogCheckpoint != nil ||
			len(value.Objects) != len(authorityObjectOrder) || previousObjects != 0 {
			return contractError("initial authority adoption receipt")
		}
	case "reaffirmation":
		if value.CatalogGeneration < 1 || value.PreviousCatalogGeneration != value.CatalogGeneration ||
			value.PreCatalogContractSHA256 != value.PostCatalogContractSHA256 ||
			!validOptionalDigest(value.PreviousAuthorityReceiptSHA256) ||
			value.UpgradeArtifactSHA256 != nil || value.PreviousCatalogCheckpoint == nil ||
			previousObjects != len(value.Objects) {
			return contractError("authority reaffirmation receipt")
		}
	case "catalog_upgrade":
		if value.PreviousCatalogGeneration < 1 ||
			value.CatalogGeneration != value.PreviousCatalogGeneration+1 ||
			value.PreCatalogContractSHA256 == value.PostCatalogContractSHA256 ||
			!validOptionalDigest(value.PreviousAuthorityReceiptSHA256) ||
			!validOptionalDigest(value.UpgradeArtifactSHA256) || value.PreviousCatalogCheckpoint == nil ||
			len(value.Objects) <= len(authorityObjectOrder) || previousObjects >= len(value.Objects) {
			return contractError("authority catalog upgrade receipt")
		}
	default:
		return contractError("authority adoption receipt mode")
	}
	return nil
}

func validateAuthorityReceiptObjects(
	objects []AuthorityAdoptionReceiptObject, generation int,
) error {
	if len(objects) < len(authorityObjectOrder) {
		return contractError("authority adoption receipt object set")
	}
	seen := make(map[string]struct{}, len(objects))
	previousGeneration := 0
	for index, object := range objects {
		if !validAuthorityReceiptIdentity(object, index) {
			return contractError("authority adoption receipt versioned object")
		}
		key := authorityObjectIdentity(
			object.Kind, object.SchemaName, object.ObjectName, object.IdentityArguments,
		)
		if _, exists := seen[key]; exists {
			return contractError("authority adoption receipt duplicate object")
		}
		seen[key] = struct{}{}
		if object.ObjectOID == 0 || object.Owner != authorityOwner ||
			object.IntroducedGeneration < 1 || object.IntroducedGeneration > generation ||
			object.IntroducedGeneration < previousGeneration ||
			!validDigest(object.DefinitionSHA256) || !validDigest(object.ACLSHA256) {
			return contractError("authority adoption receipt object")
		}
		previousGeneration = object.IntroducedGeneration
	}
	return nil
}

func (value AuthorityAdoptionSignature) Validate(contractVersion string) error {
	if value.ContractVersion != contractVersion || !validDigest(value.ManifestSHA256) {
		return contractError("authority adoption signature contract")
	}
	return validateSignature(value.Signature)
}
