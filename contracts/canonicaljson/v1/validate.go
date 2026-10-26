package canonicaljsonv1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"strconv"
	"strings"
)

var maximumSafeInteger = big.NewInt(9_007_199_254_740_991)

const maximumNestingDepth = 128

// Validate rejects ambiguous JSON before RFC 8785 canonicalization.
func Validate(document []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.UseNumber()
	if err := scanValue(decoder, 0); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}

func scanValue(decoder *json.Decoder, depth int) error {
	if depth > maximumNestingDepth {
		return fmt.Errorf("JSON nesting exceeds its limit")
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token == nil {
		return nil
	}
	if number, ok := token.(json.Number); ok {
		return validateNumber(number.String())
	}
	delimiter, composite := token.(json.Delim)
	if !composite {
		return nil
	}
	seen := map[string]struct{}{}
	for decoder.More() {
		if delimiter == '{' {
			keyToken, keyErr := decoder.Token()
			if keyErr != nil {
				return keyErr
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("invalid object key")
			}
			if _, duplicate := seen[key]; duplicate {
				return fmt.Errorf("duplicate object key")
			}
			seen[key] = struct{}{}
		}
		if err := scanValue(decoder, depth+1); err != nil {
			return err
		}
	}
	_, err = decoder.Token()
	return err
}

func validateNumber(number string) error {
	if strings.ContainsAny(number, ".eE") {
		value, err := strconv.ParseFloat(number, 64)
		if err != nil || math.IsInf(value, 0) || math.IsNaN(value) {
			return fmt.Errorf("JSON number is outside finite double precision")
		}
		return nil
	}
	integer, valid := new(big.Int).SetString(number, 10)
	if !valid || new(big.Int).Abs(integer).Cmp(maximumSafeInteger) > 0 {
		return fmt.Errorf("JSON integer exceeds the interoperable range")
	}
	return nil
}
