package profile

func RetiredField(field Field) (Field, error) {
	if !field.valid() || field.state != FieldActive {
		return Field{}, ErrFieldUnavailable
	}
	result := field.clone()
	result.state = FieldRetired
	return result, nil
}
