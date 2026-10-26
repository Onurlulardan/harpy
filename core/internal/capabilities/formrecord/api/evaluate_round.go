package recordkernel

func evaluateRound(
	decimalValueInput, scaleValue Value,
	target *DecimalProfile,
) (Value, error) {
	decimal, decimalErr := valueDecimal(decimalValueInput)
	scale, scaleErr := valueInteger(scaleValue)
	if decimalErr != nil || scaleErr != nil || target == nil || scale < 0 ||
		scale > int64(target.Scale) {
		return Value{}, ErrCalculation
	}
	_, rounded, err := roundedDecimal(decimal, int(scale), target.Precision)
	if err != nil {
		return Value{}, err
	}
	return buildRoundedDecimal(rounded, *target)
}
