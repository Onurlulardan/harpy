package config

func validateFormsArtifactKeySeparation(export FCARTKEKKeyring,
	importSource FormsImportSourceKEKKeyring) error {
	if export.Validate() != nil || importSource.Validate() != nil {
		return configError(formsImportSourceFCARTKeyringFileKey,
			"cannot prove import/export FCART key separation")
	}
	exportIDs := make(map[string]struct{}, len(export.retained)+1)
	exportMaterial := make(map[[32]byte]struct{}, len(export.retained)+1)
	for _, key := range append([]FCARTKEK{export.current}, export.retained...) {
		exportIDs[key.id], exportMaterial[key.material] = struct{}{}, struct{}{}
	}
	for _, key := range importSource.all() {
		_, duplicateID := exportIDs[key.id]
		_, duplicateMaterial := exportMaterial[key.material]
		if duplicateID || duplicateMaterial {
			return configError(formsImportSourceFCARTKeyringFileKey,
				"must not share KIDs or key material with Forms export")
		}
	}
	return nil
}
