package recordkernel

import (
	"unicode/utf8"
)

func validateFinalRecord(
	catalog Catalog,
	record Record,
	budget *instructionBudget,
) error {
	if err := validateRecordShape(catalog, record); err != nil {
		return err
	}
	issues := make([]ValidationIssue, 0)
	for _, field := range catalog.fields {
		value, present := record.values[field.ID]
		if !present || value.null {
			if field.Required || !field.Nullable {
				issues = append(issues, ValidationIssue{
					Code: ValidationRequired, FieldID: field.ID,
				})
			}
			continue
		}
		if field.TypeID == TypeText && !validTextLength(field, value) {
			issues = append(issues, ValidationIssue{
				Code: ValidationTextLength, FieldID: field.ID,
			})
		}
	}
	crossIssues, err := validateCrossFields(catalog, record, budget)
	if err != nil {
		return err
	}
	issues = append(issues, crossIssues...)
	if len(issues) > 0 {
		return newValidationFailure(issues)
	}
	return nil
}

func validTextLength(field Field, value Value) bool {
	text, err := valueString(value)
	if err != nil {
		return false
	}
	length := utf8.RuneCountInString(text)
	if field.TextMinimum != nil && length < *field.TextMinimum {
		return false
	}
	return field.TextMaximum == nil || length <= *field.TextMaximum
}
