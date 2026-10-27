package iam

type systemTransferAccessTarget struct{ action string }

func resolveSystemTransferAccess(resource SystemTransferResource,
	permission, operation string) (systemTransferAccessTarget, bool) {
	action := systemTransferAction(operation)
	valid := resource.Valid() && (action == "export" || action == "import") &&
		permission == resource.Permission(action)
	if !valid {
		return systemTransferAccessTarget{}, false
	}
	allowed := map[string]bool{
		SystemExportSubmitOperation: true, SystemExportGetOperation: true,
		SystemExportCancelOperation: true, SystemExportDownloadOperation: true,
		SystemImportSourceOperation: true, SystemImportSubmitOperation: true,
		SystemImportGetOperation: true, SystemImportCancelOperation: true,
		SystemImportReportOperation: true,
	}
	return systemTransferAccessTarget{action: action}, allowed[operation]
}
