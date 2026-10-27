package profile

type FieldState string

const (
	FieldActive  FieldState = "active"
	FieldRetired FieldState = "retired"
)

func (state FieldState) Valid() bool {
	return state == FieldActive || state == FieldRetired
}

type SchemaState string

const (
	SchemaActive  SchemaState = "active"
	SchemaRetired SchemaState = "retired"
)

func (state SchemaState) Valid() bool {
	return state == SchemaActive || state == SchemaRetired
}
