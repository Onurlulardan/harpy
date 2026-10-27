package recordkernel

func coerceExpressionValue(field Field, value Value) (Value, error) {
	if value.typeID != field.TypeID {
		return Value{}, ErrCalculation
	}
	if value.null {
		return nullValue(field.TypeID), nil
	}
	if field.TypeID == TypeDecimal {
		decimal, err := valueDecimal(value)
		if err != nil || field.Decimal == nil {
			return Value{}, ErrCalculation
		}
		return buildRoundedDecimal(decimal, *field.Decimal)
	}
	if len(value.canonical) == 0 {
		return Value{}, ErrCalculation
	}
	result, err := DecodeValue(field, value.canonical)
	if err != nil {
		return Value{}, ErrCalculation
	}
	return result, nil
}

func decimalTargetScale(field Field) int {
	if field.Decimal == nil {
		return 0
	}
	return field.Decimal.Scale
}
