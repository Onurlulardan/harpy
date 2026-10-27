package config

func validateIAMArtifactSeparation(settings IAMSystemTransferArtifacts) error {
	if settings.export.Validate() != nil || settings.importSource.Validate() != nil {
		return configError(iamImportKeyringKey, "cannot prove artifact isolation")
	}
	buckets := []string{settings.export.storage.bucket, settings.importSource.storage.bucket}
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
	rings := []FCARTKEKKeyring{settings.export.keyring, settings.importSource.keyring}
	for _, ring := range rings {
		for _, key := range append([]FCARTKEK{ring.current}, ring.retained...) {
			if err := appendKey(key.id, key.material); err != nil {
				return err
			}
		}
	}
	return nil
}
