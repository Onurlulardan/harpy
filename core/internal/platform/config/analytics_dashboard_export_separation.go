package config

func validateAnalyticsDashboardExportSeparation(
	analytics AnalyticsDashboardExportArtifacts,
	formsExport FormsExportArtifacts,
	formsImport FormsImportSourceArtifacts,
	iam IAMSystemTransferArtifacts,
) error {
	if analytics.Validate() != nil || formsExport.Validate() != nil ||
		formsImport.Validate() != nil || iam.Validate() != nil {
		return configError(analyticsExportKeyringKey,
			"cannot prove artifact authority separation")
	}
	for _, bucket := range []string{
		formsExport.storage.bucket, formsImport.storage.bucket,
		iam.export.storage.bucket, iam.importSource.storage.bucket,
	} {
		if analytics.storage.bucket == bucket {
			return configError(analyticsExportBucketKey,
				"must not share a bucket with another artifact authority")
		}
	}
	for _, candidate := range append([]FCARTKEK{analytics.keyring.current},
		analytics.keyring.retained...) {
		for _, ring := range []FCARTKEKKeyring{
			formsExport.keyring, iam.export.keyring, iam.importSource.keyring,
		} {
			for _, existing := range append([]FCARTKEK{ring.current}, ring.retained...) {
				if candidate.id == existing.id || candidate.material == existing.material {
					return configError(analyticsExportKeyringKey,
						"must not share KIDs or key material with another artifact authority")
				}
			}
		}
		for _, existing := range formsImport.keyring.all() {
			if candidate.id == existing.id || candidate.material == existing.material {
				return configError(analyticsExportKeyringKey,
					"must not share KIDs or key material with another artifact authority")
			}
		}
	}
	return nil
}
