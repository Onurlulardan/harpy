package rowdecode

import (
	"bytes"
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/jackc/pgx/v5/pgtype"
)

func FuzzDecodeTextNeverPanicsOrLeaksFailureInput(f *testing.F) {
	for _, seed := range []string{"", "hello", "Caf\u00e9", "e\u0301", "a\x00b"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		value, err := Decode(textType(t), input)
		if err != nil {
			if input != "" && strings.Contains(err.Error(), input) {
				t.Fatal("failure leaked input")
			}
			return
		}
		response, responseErr := value.ResponseJSON()
		last, present, lastErr := value.LastValue()
		if responseErr != nil || !json.Valid(response) || lastErr != nil || !present ||
			last.Validate() != nil {
			t.Fatal("successful text decode produced invalid output")
		}
	})
}

func FuzzDecodeJSONIsStrictAndCanonical(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`{"a":1}`), []byte(`{"a":1,"a":2}`), []byte(`null`), []byte(`[`),
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input []byte) {
		fieldType, _ := api.NewJSONFieldType()
		value, err := Decode(fieldType, input)
		if err != nil {
			return
		}
		response, responseErr := value.ResponseJSON()
		canonical, canonicalErr := attestation.Canonicalize(response)
		if responseErr != nil || canonicalErr != nil || !bytes.Equal(response, canonical) {
			t.Fatal("successful JSON decode was not canonical")
		}
	})
}

func FuzzDecodeDecimalBoundsExponentAndPrecision(f *testing.F) {
	f.Add(int64(1234), int32(-2))
	f.Add(int64(1), int32(1000))
	f.Add(int64(0), int32(-1000))
	f.Fuzz(func(t *testing.T, coefficient int64, exponent int32) {
		value, err := Decode(decimalType(t, 18, 6), pgtype.Numeric{
			Int: big.NewInt(coefficient), Exp: exponent, Valid: true,
		})
		if err != nil {
			return
		}
		if value.Validate() != nil {
			t.Fatal("successful decimal decode produced invalid value")
		}
	})
}
