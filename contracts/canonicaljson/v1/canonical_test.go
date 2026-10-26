package canonicaljsonv1

import (
	"bytes"
	"testing"
)

func TestCanonicalizeUsesRFC8785AndIsIdempotent(t *testing.T) {
	t.Parallel()
	input := []byte(`{"numbers":[333333333.33333329,1E30,4.50,2e-3],"z":0,"a":1}`)
	want := `{"a":1,"numbers":[333333333.3333333,1e+30,4.5,0.002],"z":0}`
	canonical, err := CanonicalizeObject(input, 1024)
	if err != nil {
		t.Fatal(err)
	}
	if string(canonical) != want {
		t.Fatalf("canonical JSON = %s", canonical)
	}
	again, err := CanonicalizeObject(canonical, 1024)
	if err != nil || !bytes.Equal(again, canonical) {
		t.Fatalf("second canonicalization = %s, %v", again, err)
	}
}

func TestCanonicalizeRejectsAmbiguousOrUnboundedInput(t *testing.T) {
	t.Parallel()
	tests := [][]byte{
		{}, []byte(`{"a":1,"a":2}`), []byte(`{"a":1} {"b":2}`),
		[]byte(`"\ud800"`),
		[]byte(`{"value":9007199254740992}`), []byte(`{"value":1e400}`),
		append([]byte{'"'}, 0xff, '"'),
	}
	for _, document := range tests {
		if _, err := Canonicalize(document, 64); err == nil {
			t.Fatalf("ambiguous JSON accepted: %q", document)
		}
	}
	if _, err := Canonicalize(bytes.Repeat([]byte{' '}, 65), 64); err == nil {
		t.Fatal("oversized JSON was accepted")
	}
}

func TestCanonicalizeAcceptsInteroperableBoundaryValues(t *testing.T) {
	t.Parallel()
	document := []byte(`{"null":null,"safe":9007199254740991}`)
	canonical, err := CanonicalizeObject(document, 128)
	if err != nil {
		t.Fatal(err)
	}
	if string(canonical) != `{"null":null,"safe":9007199254740991}` {
		t.Fatalf("canonical boundary JSON = %s", canonical)
	}
}

func TestCanonicalizeObjectRejectsNonObjectRoot(t *testing.T) {
	t.Parallel()
	for _, document := range [][]byte{[]byte(`[]`), []byte(`null`), []byte(`"x"`)} {
		if _, err := CanonicalizeObject(document, 64); err == nil {
			t.Fatalf("non-object root accepted: %s", document)
		}
	}
}

func TestCanonicalizeRejectsExcessiveNesting(t *testing.T) {
	t.Parallel()
	document := append(bytes.Repeat([]byte{'['}, maximumNestingDepth+1), '0')
	document = append(document, bytes.Repeat([]byte{']'}, maximumNestingDepth+1)...)
	if _, err := Canonicalize(document, len(document)); err == nil {
		t.Fatal("excessively nested JSON was accepted")
	}
}
