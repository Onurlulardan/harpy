package recordkernel

import (
	"encoding/json"
	"math/big"
	"strconv"
)

func valueString(value Value) (string, error) {
	var result string
	if value.null || json.Unmarshal(value.canonical, &result) != nil {
		return "", ErrCalculation
	}
	return result, nil
}

func valueInteger(value Value) (int64, error) {
	text, err := valueString(value)
	if err != nil {
		return 0, err
	}
	result, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return 0, ErrCalculation
	}
	return result, nil
}

func valueBoolean(value Value) (bool, error) {
	if value.null {
		return false, ErrCalculation
	}
	if string(value.canonical) == "true" {
		return true, nil
	}
	if string(value.canonical) == "false" {
		return false, nil
	}
	return false, ErrCalculation
}

func valueDecimal(value Value) (*big.Rat, error) {
	if value.null {
		return nil, ErrCalculation
	}
	if value.decimal != nil {
		return new(big.Rat).Set(value.decimal), nil
	}
	text, err := valueString(value)
	if err != nil {
		return nil, err
	}
	result, valid := new(big.Rat).SetString(text)
	if !valid {
		return nil, ErrCalculation
	}
	return result, nil
}
