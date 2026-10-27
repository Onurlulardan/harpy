package api

import "fmt"

func (fieldType FieldType) Kind() (LogicalType, error) {
	if err := fieldType.Validate(); err != nil {
		return 0, err
	}
	return fieldType.kind, nil
}

func (fieldType FieldType) Text() (TextProfile, bool, error) {
	if err := fieldType.Validate(); err != nil {
		return TextProfile{}, false, err
	}
	switch state := fieldType.state.(type) {
	case textFieldState:
		return state.profile, true, nil
	case textListFieldState:
		return state.profile, true, nil
	default:
		return TextProfile{}, false, nil
	}
}

func (fieldType FieldType) Integer() (IntegerProfile, bool, error) {
	if err := fieldType.Validate(); err != nil {
		return IntegerProfile{}, false, err
	}
	state, valid := fieldType.state.(integerFieldState)
	return state.profile, valid, nil
}

func (fieldType FieldType) Decimal() (DecimalProfile, bool, error) {
	if err := fieldType.Validate(); err != nil {
		return DecimalProfile{}, false, err
	}
	state, valid := fieldType.state.(decimalFieldState)
	return state.profile, valid, nil
}

func (fieldType FieldType) Enum() (EnumProfile, bool, error) {
	if err := fieldType.Validate(); err != nil {
		return EnumProfile{}, false, err
	}
	state, valid := fieldType.state.(enumFieldState)
	if !valid {
		return EnumProfile{}, false, nil
	}
	return state.profile.clone(), true, nil
}

func (fieldType FieldType) String() string   { return queryValueRedaction }
func (fieldType FieldType) GoString() string { return queryValueRedaction }
func (fieldType FieldType) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (FieldType) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
