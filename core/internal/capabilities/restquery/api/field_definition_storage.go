package api

func fieldStorageMatchesType(fieldType FieldType, storage FieldStorage) bool {
	kind, err := fieldType.Kind()
	if err != nil {
		return false
	}
	_, hasFolded, foldedErr := storage.FoldedCompanion()
	_, hasAccent, accentErr := storage.AccentCompanion()
	if foldedErr != nil || accentErr != nil {
		return false
	}
	if kind != LogicalTypeText {
		return !hasFolded && !hasAccent
	}
	profile, found, profileErr := fieldType.Text()
	if profileErr != nil || !found {
		return false
	}
	_, requiresFolded, casefoldErr := profile.Casefold()
	_, requiresAccent, accentProfileErr := profile.Accent()
	return casefoldErr == nil && accentProfileErr == nil &&
		hasFolded == requiresFolded && hasAccent == requiresAccent
}
