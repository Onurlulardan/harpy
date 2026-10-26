package recordkernel

func validateRecordShape(catalog Catalog, record Record) error {
	issues := make([]ValidationIssue, 0)
	for fieldID, value := range record.values {
		field, found := catalog.fieldsByID[fieldID]
		if !found {
			issues = append(issues, ValidationIssue{
				Code: ValidationUnknown, FieldID: fieldID,
			})
			continue
		}
		if value.typeID != field.TypeID || len(value.canonical) > MaximumFieldBytes {
			issues = append(issues, ValidationIssue{
				Code: ValidationType, FieldID: fieldID,
			})
			continue
		}
		if value.null && string(value.canonical) == "null" {
			continue
		}
		normalized, _, err := decodeInputValue(field, value.canonical)
		if err != nil || !valuesEqual(normalized, value) {
			issues = append(issues, ValidationIssue{
				Code: ValidationType, FieldID: fieldID,
			})
		}
	}
	if len(issues) > 0 {
		return newValidationFailure(issues)
	}
	canonical, err := canonicalFieldIDDocument(record.values)
	if err != nil || uint64(len(canonical)) > catalog.limits.RecordInputBytes ||
		len(canonical) > MaximumRecordBytes {
		return ErrBudgetExceeded
	}
	return nil
}
