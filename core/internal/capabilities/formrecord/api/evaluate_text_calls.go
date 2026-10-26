package recordkernel

import (
	"strings"
)

func evaluateConcat(values []Value) (Value, error) {
	var builder strings.Builder
	for _, value := range values {
		text, err := valueString(value)
		if err != nil {
			return Value{}, err
		}
		if builder.Len()+len(text) > MaximumFieldBytes {
			return Value{}, ErrBudgetExceeded
		}
		builder.WriteString(text)
	}
	return buildText(builder.String())
}

func evaluateTextCase(value Value, upper bool) (Value, error) {
	text, err := valueString(value)
	if err != nil {
		return Value{}, err
	}
	if upper {
		text = strings.ToUpper(text)
	} else {
		text = strings.ToLower(text)
	}
	return buildText(text)
}
