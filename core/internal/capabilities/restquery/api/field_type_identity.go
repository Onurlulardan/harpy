package api

import (
	"fmt"
	"slices"
)

// Clone returns a defensive copy of the exact declared type profile.
func (fieldType FieldType) Clone() (FieldType, error) {
	if err := fieldType.Validate(); err != nil {
		return FieldType{}, fmt.Errorf("REST Query field type is invalid")
	}
	return fieldType.clone(), nil
}

// ExactEqual compares the complete type contract, not only its logical kind.
func (fieldType FieldType) ExactEqual(other FieldType) (bool, error) {
	if fieldType.Validate() != nil || other.Validate() != nil {
		return false, fmt.Errorf("REST Query field type is invalid")
	}
	if fieldType.kind != other.kind {
		return false, nil
	}
	switch left := fieldType.state.(type) {
	case textFieldState:
		right, ok := other.state.(textFieldState)
		return ok && left.profile == right.profile, nil
	case integerFieldState:
		right, ok := other.state.(integerFieldState)
		return ok && left.profile == right.profile, nil
	case decimalFieldState:
		right, ok := other.state.(decimalFieldState)
		return ok && left.profile == right.profile, nil
	case enumFieldState:
		right, ok := other.state.(enumFieldState)
		return ok && equalEnumProfiles(left.profile, right.profile), nil
	case textListFieldState:
		right, ok := other.state.(textListFieldState)
		return ok && left.profile == right.profile, nil
	case fixedFieldState:
		right, ok := other.state.(fixedFieldState)
		return ok && left == right, nil
	default:
		return false, fmt.Errorf("REST Query field type is invalid")
	}
}

func equalEnumProfiles(left, right EnumProfile) bool {
	return left.version == right.version && left.ordered == right.ordered &&
		left.validated == right.validated && slices.Equal(left.values, right.values)
}
