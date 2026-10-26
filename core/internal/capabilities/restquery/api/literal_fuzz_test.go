package api

import (
	"bytes"
	"encoding/json"
	"testing"
)

func FuzzLiteralNeverRetainsCallerStorage(f *testing.F) {
	f.Add([]byte(`"text"`))
	f.Add([]byte(`{"a":[1,true]}`))
	f.Add([]byte("null"))
	f.Fuzz(func(t *testing.T, input []byte) {
		original := bytes.Clone(input)
		literal, err := NewJSONLiteral(input)
		if err != nil {
			return
		}
		if len(input) > 0 {
			input[0] ^= 0xff
		}
		first, present, accessorErr := literal.JSONValue()
		if accessorErr != nil || !present {
			t.Fatal("accepted JSON literal became unavailable")
		}
		if !json.Valid(first) || bytes.Equal(first, []byte("null")) {
			t.Fatalf("accepted literal became invalid; input=%q", original)
		}
		if len(first) > 0 {
			first[0] ^= 0xff
		}
		second, _, _ := literal.JSONValue()
		if !json.Valid(second) {
			t.Fatal("output mutation changed literal")
		}
	})
}
