package runtimev1

func addReleaseVerificationSnapshot(variables map[string]any) {
	addReleaseInputsSnapshot(variables)
	variables["HARPY_BOOTSTRAP_RECEIPT_PATH"] = "/receipts/bootstrap.json"
	variables["HARPY_EXTENSION_RECEIPT_PATH"] = "/receipts/extension.json"
	variables["HARPY_CREDENTIAL_CLEANUP_RECEIPT_PATH"] = "/receipts/cleanup.json"
	variables["HARPY_CATALOG_PACKAGE_SET_PATH"] = "/release/catalog-package-set.json"
	variables["HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING"] =
		validCatalogPackageKeyringContractValue()
	variables["HARPY_CUTOVER_AUTHORITY_OPERATION_PATH"] =
		"/authority/cutover-operation.json"
	variables["HARPY_CUTOVER_AUTHORITY_OPERATION_SIGNATURE_PATH"] =
		"/authority/cutover-operation.sig"
	variables["HARPY_CUTOVER_AUTHORITY_RECEIPT_PATH"] =
		"/authority/cutover-receipt.json"
	variables["HARPY_CUTOVER_AUTHORITY_RECEIPT_SIGNATURE_PATH"] =
		"/authority/cutover-receipt.sig"
	addAuthorityArtifactsSnapshot(variables)
	addMigrationArtifactsSnapshot(variables)
}
