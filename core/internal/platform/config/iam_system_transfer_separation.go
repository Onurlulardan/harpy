package config

func validateIAMArtifactSeparation(settings IAMSystemTransferArtifacts,
	formsExport FormsExportArtifacts, formsImport FormsImportSourceArtifacts) error {
	if formsExport.Validate() != nil || formsImport.Validate() != nil ||
		settings.export.Validate() != nil || settings.importSource.Validate() != nil {
		return configError(iamImportKeyringKey, "cannot prove artifact isolation")
	}
	buckets := []string{formsExport.storage.bucket, formsImport.storage.bucket,
		settings.export.storage.bucket, settings.importSource.storage.bucket}
	seenBuckets := make(map[string]struct{}, len(buckets))
	for _, bucket := range buckets {
		if _, exists := seenBuckets[bucket]; exists {
			return configError(iamImportBucketKey,
				"must not share a bucket with another artifact authority")
		}
		seenBuckets[bucket] = struct{}{}
	}
	seenIDs := map[string]struct{}{}
	seenMaterial := map[[32]byte]struct{}{}
	appendKey := func(id string, material [32]byte) error {
		_, duplicateID := seenIDs[id]
		_, duplicateMaterial := seenMaterial[material]
		if duplicateID || duplicateMaterial {
			return configError(iamImportKeyringKey,
				"must not share KIDs or key material with another artifact authority")
		}
		seenIDs[id], seenMaterial[material] = struct{}{}, struct{}{}
		return nil
	}
	for _, ring := range []FCARTKEKKeyring{
		formsExport.keyring, settings.export.keyring, settings.importSource.keyring,
	} {
		for _, key := range append([]FCARTKEK{ring.current}, ring.retained...) {
			if err := appendKey(key.id, key.material); err != nil {
				return err
			}
		}
	}
	for _, key := range formsImport.keyring.all() {
		if err := appendKey(key.id, key.material); err != nil {
			return err
		}
	}
	return nil
}
