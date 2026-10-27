package config

func validateAnalyticsDashboardExportSeparation(
	analytics AnalyticsDashboardExportArtifacts,
	formsExport FormsExportArtifacts,
	formsImport FormsImportSourceArtifacts,
	iam IAMSystemTransferArtifacts,
) error {
	formsActive := formsExport.valid && formsImport.valid
	if analytics.Validate() != nil || iam.Validate() != nil ||
		formsExport.valid != formsImport.valid ||
		(formsActive && (formsExport.Validate() != nil || formsImport.Validate() != nil)) {
		return configError(analyticsExportKeyringKey,
			"cannot prove artifact authority separation")
	}
	buckets := []string{iam.export.storage.bucket, iam.importSource.storage.bucket}
	rings := []FCARTKEKKeyring{iam.export.keyring, iam.importSource.keyring}
	if formsActive {
		buckets = append(buckets, formsExport.storage.bucket, formsImport.storage.bucket)
		rings = append(rings, formsExport.keyring)
	}
	for _, bucket := range buckets {
		if analytics.storage.bucket == bucket {
			return configError(analyticsExportBucketKey,
				"must not share a bucket with another artifact authority")
		}
	}
	for _, candidate := range append([]FCARTKEK{analytics.keyring.current},
		analytics.keyring.retained...) {
		for _, ring := range rings {
			for _, existing := range append([]FCARTKEK{ring.current}, ring.retained...) {
				if candidate.id == existing.id || candidate.material == existing.material {
					return configError(analyticsExportKeyringKey,
						"must not share KIDs or key material with another artifact authority")
				}
			}
		}
		if !formsActive {
			continue
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
