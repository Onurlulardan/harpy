package api

type FieldType struct {
	kind      LogicalType
	state     fieldTypeState
	validated bool
}

func NewTextFieldType(profile TextProfile) (FieldType, error) {
	return newFieldType(LogicalTypeText, textFieldState{profile: profile})
}

func NewIntegerFieldType(profile IntegerProfile) (FieldType, error) {
	return newFieldType(LogicalTypeInteger, integerFieldState{profile: profile})
}

func NewDecimalFieldType(profile DecimalProfile) (FieldType, error) {
	return newFieldType(LogicalTypeDecimal, decimalFieldState{profile: profile})
}

func NewBooleanFieldType() (FieldType, error) {
	return newFixedFieldType(LogicalTypeBoolean)
}

func NewDateFieldType() (FieldType, error) {
	return newFixedFieldType(LogicalTypeDate)
}

func NewTimestampFieldType() (FieldType, error) {
	return newFixedFieldType(LogicalTypeTimestamp)
}

func NewUUIDFieldType() (FieldType, error) {
	return newFixedFieldType(LogicalTypeUUID)
}

func NewEnumFieldType(profile EnumProfile) (FieldType, error) {
	return newFieldType(LogicalTypeEnum, enumFieldState{profile: profile.clone()})
}

func NewRelationFieldType() (FieldType, error) {
	return newFixedFieldType(LogicalTypeRelation)
}

func NewTextListFieldType(profile TextProfile) (FieldType, error) {
	return newFieldType(LogicalTypeTextList, textListFieldState{profile: profile})
}

func NewUUIDListFieldType() (FieldType, error) {
	return newFixedFieldType(LogicalTypeUUIDList)
}

func NewJSONFieldType() (FieldType, error) {
	return newFixedFieldType(LogicalTypeJSON)
}

func newFixedFieldType(kind LogicalType) (FieldType, error) {
	return newFieldType(kind, fixedFieldState{kind: kind})
}

func newFieldType(kind LogicalType, state fieldTypeState) (FieldType, error) {
	fieldType := FieldType{kind: kind, state: state, validated: true}
	if err := fieldType.Validate(); err != nil {
		return FieldType{}, err
	}
	return fieldType, nil
}
