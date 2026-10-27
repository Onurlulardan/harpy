package config

func formsProductConfigured(values environmentValues) bool {
	keys := append(formsExportArtifactKeys(), formsImportSourceArtifactKeys()...)
	keys = append(keys, formsImportSourceStagingKey, "HARPY_FORMS_EXPORT_SPOOL_DIR")
	for _, key := range keys {
		if _, found := values.lookup(key); found {
			return true
		}
	}
	return false
}

func skipFormsProduct(environment Environment, values environmentValues) bool {
	return !environment.productionLike() && !formsProductConfigured(values)
}

func loadOptionalFormsExport(
	values environmentValues,
	environment Environment,
) (FormsExportArtifacts, error) {
	if skipFormsProduct(environment, values) {
		return FormsExportArtifacts{}, nil
	}
	return loadFormsExportArtifacts(values, environment)
}

func loadOptionalFormsImport(
	values environmentValues,
	environment Environment,
	export FormsExportArtifacts,
) (FormsImportSourceArtifacts, error) {
	if skipFormsProduct(environment, values) {
		return FormsImportSourceArtifacts{}, nil
	}
	return loadFormsImportSourceArtifacts(values, environment, export)
}

func loadOptionalFormsStaging(
	values environmentValues,
	environment Environment,
) (AbsolutePath, error) {
	if skipFormsProduct(environment, values) {
		return AbsolutePath{}, nil
	}
	return loadAbsolutePath(values, formsImportSourceStagingKey)
}

func loadOptionalFormsSpool(
	values environmentValues,
	environment Environment,
) (AbsolutePath, error) {
	if skipFormsProduct(environment, values) {
		return AbsolutePath{}, nil
	}
	return loadAbsolutePath(values, "HARPY_FORMS_EXPORT_SPOOL_DIR")
}
