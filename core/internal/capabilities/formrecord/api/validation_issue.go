package recordkernel

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
)

const (
	ValidationRequired   = "FORM_RECORD_REQUIRED"
	ValidationTextLength = "FORM_RECORD_TEXT_LENGTH"
	ValidationCrossField = "FORM_RECORD_CROSS_FIELD"
	ValidationType       = "FORM_RECORD_TYPE"
	ValidationUnknown    = "FORM_RECORD_UNKNOWN_FIELD"
)

type ValidationIssue struct {
	Code    string
	FieldID uuid.UUID
	RuleID  uuid.UUID
}

type ValidationFailure struct {
	issues []ValidationIssue
}

func (failure ValidationFailure) Error() string {
	return fmt.Sprintf("%s: %d issue(s)", ErrValidation.Error(), len(failure.issues))
}

func (failure ValidationFailure) Unwrap() error { return ErrValidation }

func (failure ValidationFailure) Issues() []ValidationIssue {
	return slices.Clone(failure.issues)
}

func newValidationFailure(issues []ValidationIssue) error {
	slices.SortFunc(issues, compareValidationIssue)
	return ValidationFailure{issues: slices.Clone(issues)}
}

func compareValidationIssue(left, right ValidationIssue) int {
	if compared := compareUUID(left.FieldID, right.FieldID); compared != 0 {
		return compared
	}
	if compared := compareUUID(left.RuleID, right.RuleID); compared != 0 {
		return compared
	}
	return compareText(left.Code, right.Code)
}
