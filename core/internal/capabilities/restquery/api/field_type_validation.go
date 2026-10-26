package api

import "fmt"

type fieldTypeState interface {
	fieldTypeState()
}

type textFieldState struct{ profile TextProfile }
type integerFieldState struct{ profile IntegerProfile }
type decimalFieldState struct{ profile DecimalProfile }
type enumFieldState struct{ profile EnumProfile }
type textListFieldState struct{ profile TextProfile }
type fixedFieldState struct{ kind LogicalType }

func (textFieldState) fieldTypeState()     {}
func (integerFieldState) fieldTypeState()  {}
func (decimalFieldState) fieldTypeState()  {}
func (enumFieldState) fieldTypeState()     {}
func (textListFieldState) fieldTypeState() {}
func (fixedFieldState) fieldTypeState()    {}

func (fieldType FieldType) Validate() error {
	if !fieldType.validated || !fieldType.kind.Valid() || !fieldType.validState() {
		return fmt.Errorf("REST Query field type is invalid")
	}
	return nil
}

func (fieldType FieldType) validState() bool {
	switch fieldType.kind {
	case LogicalTypeText:
		state, valid := fieldType.state.(textFieldState)
		return valid && state.profile.Validate() == nil
	case LogicalTypeInteger:
		state, valid := fieldType.state.(integerFieldState)
		return valid && state.profile.Validate() == nil
	case LogicalTypeDecimal:
		state, valid := fieldType.state.(decimalFieldState)
		return valid && state.profile.Validate() == nil
	case LogicalTypeEnum:
		state, valid := fieldType.state.(enumFieldState)
		return valid && state.profile.Validate() == nil
	case LogicalTypeTextList:
		state, valid := fieldType.state.(textListFieldState)
		return valid && exactOnlyTextProfile(state.profile)
	default:
		state, valid := fieldType.state.(fixedFieldState)
		return valid && state.kind == fieldType.kind && fixedLogicalType(fieldType.kind)
	}
}

func exactOnlyTextProfile(profile TextProfile) bool {
	return profile.Validate() == nil && profile.casefold == TextCasefoldDisabled &&
		profile.accent == TextAccentDisabled
}

func fixedLogicalType(kind LogicalType) bool {
	switch kind {
	case LogicalTypeBoolean, LogicalTypeDate, LogicalTypeTimestamp,
		LogicalTypeUUID, LogicalTypeRelation, LogicalTypeUUIDList, LogicalTypeJSON:
		return true
	default:
		return false
	}
}

func (fieldType FieldType) clone() FieldType {
	result := fieldType
	if state, valid := fieldType.state.(enumFieldState); valid {
		result.state = enumFieldState{profile: state.profile.clone()}
	}
	return result
}
