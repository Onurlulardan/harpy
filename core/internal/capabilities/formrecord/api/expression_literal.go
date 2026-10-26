package recordkernel

import (
	"bytes"
	"encoding/json"
	"math/big"
	"strings"
)

func decodeExpressionLiteral(typeID TypeID, raw []byte) (Value, error) {
	if bytes.Equal(raw, []byte("null")) {
		return nullValue(typeID), nil
	}
	if typeID == TypeDecimal {
		return decodeGenericDecimal(raw)
	}
	field := Field{TypeID: typeID, Nullable: true}
	return DecodeValue(field, raw)
}

func decodeGenericDecimal(raw []byte) (Value, error) {
	var text string
	if json.Unmarshal(raw, &text) != nil || !decimalPattern.MatchString(text) ||
		negativeZero(text) {
		return Value{}, ErrCatalogInvalid
	}
	digits := strings.ReplaceAll(strings.TrimPrefix(text, "-"), ".", "")
	if len(digits) > 1000 {
		return Value{}, ErrCatalogInvalid
	}
	encoded, _ := json.Marshal(text)
	result := typedValue(TypeDecimal, encoded)
	result.decimal, _ = new(big.Rat).SetString(text)
	return result, nil
}

func legacyLiteral(typeID TypeID, raw []byte) (Value, error) {
	if typeID == TypeInteger || typeID == TypeDecimal {
		var number json.Number
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		if decoder.Decode(&number) == nil {
			encoded, _ := json.Marshal(number.String())
			return decodeExpressionLiteral(typeID, encoded)
		}
	}
	return decodeExpressionLiteral(typeID, raw)
}
