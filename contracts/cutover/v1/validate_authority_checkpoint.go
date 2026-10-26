package cutoverv1

func validateAuthorityCatalogCheckpoint(receipt AuthorityAdoptionReceipt) error {
	checkpoint := receipt.PreviousCatalogCheckpoint
	if receipt.Mode == "initial_adoption" {
		if checkpoint != nil {
			return contractError("initial authority checkpoint")
		}
		return nil
	}
	if checkpoint == nil || receipt.PreviousAuthorityReceiptSHA256 == nil ||
		!validDigest(checkpoint.ReceiptSHA256) || !validUUID(checkpoint.ReceiptID) ||
		!validUUID(checkpoint.EnvironmentID) || ValidateDatabaseIdentity(checkpoint.Target) != nil ||
		checkpoint.CatalogGeneration != receipt.PreviousCatalogGeneration ||
		checkpoint.ReceiptSHA256 != *receipt.PreviousAuthorityReceiptSHA256 ||
		checkpoint.EnvironmentID != receipt.EnvironmentID || checkpoint.Target != receipt.Target ||
		checkpoint.CatalogContractSHA256 != receipt.PreCatalogContractSHA256 ||
		checkpoint.CatalogContractSHA256 !=
			AuthorityReceiptCatalogContractSHA256(checkpoint.Objects) ||
		checkpoint.CatalogInstanceSHA256 !=
			AuthorityReceiptCatalogInstanceSHA256(checkpoint.Objects) {
		return contractError("authority catalog checkpoint binding")
	}
	if err := validateAuthorityReceiptObjects(
		checkpoint.Objects, checkpoint.CatalogGeneration,
	); err != nil {
		return contractError("authority catalog checkpoint objects")
	}
	currentAt, currentErr := parseUTC(receipt.AdoptedAt)
	previousAt, previousErr := parseUTC(checkpoint.AdoptedAt)
	if currentErr != nil || previousErr != nil || !currentAt.After(previousAt) ||
		receipt.PreAdoptionOwnedObjectCount != len(checkpoint.Objects) ||
		len(receipt.Objects) < len(checkpoint.Objects) {
		return contractError("authority catalog checkpoint continuity")
	}
	for index, object := range checkpoint.Objects {
		if receipt.Objects[index] != object {
			return contractError("authority catalog checkpoint retained object")
		}
	}
	if receipt.Mode == "reaffirmation" && len(receipt.Objects) != len(checkpoint.Objects) {
		return contractError("authority reaffirmation checkpoint size")
	}
	if receipt.Mode == "catalog_upgrade" && len(receipt.Objects) <= len(checkpoint.Objects) {
		return contractError("authority upgrade checkpoint size")
	}
	return nil
}
