package tabularmapping

import (
	"encoding/json"
	"math/big"
)

const maximumNumericBits = 16_384

func evaluateNumeric(
	node numericNode,
	source map[string]json.RawMessage,
) (string, bool, error) {
	value, missing, err := evaluateNumericRat(node, source)
	if err != nil || missing {
		return "", missing, err
	}
	text, err := finiteDecimal(value)
	if err != nil || len(text) > MaximumNumericTextBytes {
		return "", false, ErrInvalid
	}
	return text, false, nil
}

func evaluateNumericRat(
	node numericNode,
	source map[string]json.RawMessage,
) (*big.Rat, bool, error) {
	if node.Kind == "literal" {
		text, _ := numericText(node.Value)
		return parseRat(text)
	}
	if node.Kind == "source" {
		raw, found := source[node.Source]
		if !found || string(raw) == "null" {
			return nil, true, nil
		}
		text, ok := numericText(raw)
		if !ok || len(text) > MaximumSourceBytes || !decimalPattern.MatchString(text) {
			return nil, false, ErrInvalid
		}
		return parseRat(text)
	}
	left, missing, err := evaluateNumericRat(node.Arguments[0], source)
	if err != nil || missing {
		return nil, missing, err
	}
	if node.Kind == "negate" {
		return boundedRat(new(big.Rat).Neg(left))
	}
	right, missing, err := evaluateNumericRat(node.Arguments[1], source)
	if err != nil || missing {
		return nil, missing, err
	}
	result := new(big.Rat)
	switch node.Kind {
	case "add":
		result.Add(left, right)
	case "subtract":
		result.Sub(left, right)
	case "multiply":
		result.Mul(left, right)
	case "divide":
		if right.Sign() == 0 {
			return nil, false, ErrInvalid
		}
		result.Quo(left, right)
	default:
		return nil, false, ErrInvalid
	}
	return boundedRat(result)
}

func numericText(raw json.RawMessage) (string, bool) {
	if len(raw) < 1 {
		return "", false
	}
	var text string
	if raw[0] == '"' {
		return text, json.Unmarshal(raw, &text) == nil
	}
	return string(raw), true
}

func parseRat(text string) (*big.Rat, bool, error) {
	value, ok := new(big.Rat).SetString(text)
	if !ok {
		return nil, false, ErrInvalid
	}
	return boundedRat(value)
}

func boundedRat(value *big.Rat) (*big.Rat, bool, error) {
	if value == nil || value.Num().BitLen() > maximumNumericBits ||
		value.Denom().BitLen() > maximumNumericBits {
		return nil, false, ErrInvalid
	}
	return value, false, nil
}
