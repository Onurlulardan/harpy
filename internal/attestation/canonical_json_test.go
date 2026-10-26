package attestation

import (
	"bytes"
	"testing"
)

func TestCanonicalizeUsesRFC8785OrderingAndNumbers(t *testing.T) {
	canonical, err := Canonicalize([]byte(`{"z":0,"number":333333333.33333329,"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	expected := `{"a":1,"number":333333333.3333333,"z":0}`
	if string(canonical) != expected {
		t.Fatalf("canonical JSON = %s", canonical)
	}
}

func TestDigestHexIgnoresJSONMemberOrder(t *testing.T) {
	first, err := DigestHex([]byte(`{"b":2,"a":1}`))
	if err != nil {
		t.Fatal(err)
	}
	second, err := DigestHex([]byte("{\n  \"a\": 1, \"b\": 2\n}"))
	if err != nil {
		t.Fatal(err)
	}
	if first != second || len(first) != 64 {
		t.Fatalf("digests differ: %q %q", first, second)
	}
}

func TestCanonicalizeRejectsAmbiguousInput(t *testing.T) {
	tests := [][]byte{
		[]byte(`{"a":1,"a":2}`),
		[]byte(`{"a":1} {"b":2}`),
		[]byte(`"\ud800"`),
		[]byte(`{"value":9007199254740992}`),
		[]byte(`{"value":9007199254740993}`),
		[]byte(`{"value":1e400}`),
		append([]byte{'"'}, 0xff, '"'),
		bytes.Repeat([]byte{' '}, maximumDocumentBytes+1),
	}
	for _, input := range tests {
		if _, err := Canonicalize(input); err == nil {
			t.Fatalf("ambiguous JSON accepted: %q", input)
		}
	}
}

func TestCanonicalizeAcceptsMaximumSafeInteger(t *testing.T) {
	canonical, err := Canonicalize([]byte(`{"value":9007199254740991}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(canonical) != `{"value":9007199254740991}` {
		t.Fatalf("unexpected safe integer canonical form: %s", canonical)
	}
}

func TestCanonicalizeAcceptsRFC8785Null(t *testing.T) {
	canonical, err := Canonicalize([]byte(`{"value":null}`))
	if err != nil {
		t.Fatal(err)
	}
	if string(canonical) != `{"value":null}` {
		t.Fatalf("unexpected null canonical form: %s", canonical)
	}
}
