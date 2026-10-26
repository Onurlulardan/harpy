package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestLiteralAcceptsOneJSONValueAndOwnsBytes(t *testing.T) {
	input := []byte(" \n{\"field\":[1,true,null]}\t")
	literal, err := NewJSONLiteral(input)
	if err != nil {
		t.Fatal(err)
	}
	input[3] = 'X'
	want := []byte("{\"field\":[1,true,null]}")
	value, present, accessorErr := literal.JSONValue()
	if accessorErr != nil || !present || !bytes.Equal(value, want) {
		t.Fatalf("literal retained input storage: %s", value)
	}
	output, _, _ := literal.JSONValue()
	output[0] = 'X'
	second, _, _ := literal.JSONValue()
	if !bytes.Equal(second, want) {
		t.Fatal("literal exposed mutable storage")
	}
}

func TestTextLiteralPreservesDecodedLeafWithoutGuessingType(t *testing.T) {
	for _, input := range []string{"", "true", "1000", "Cafe\u0301", `{"a":1}`} {
		literal, err := NewTextLiteral(input)
		if err != nil {
			t.Fatal(err)
		}
		value, present, accessorErr := literal.TextValue()
		if accessorErr != nil || !present || value != input {
			t.Fatalf("text literal = %q, present = %v", value, present)
		}
		if _, present, accessorErr := literal.JSONValue(); accessorErr != nil || present {
			t.Fatal("text literal was reported as POST JSON")
		}
	}
	for _, input := range []string{"bad\x00value", string([]byte{0xff})} {
		if _, err := NewTextLiteral(input); err == nil {
			t.Fatal("invalid text literal accepted")
		}
	}
}

func TestLiteralRejectsInvalidJSONBoundaries(t *testing.T) {
	tests := [][]byte{
		nil,
		{},
		[]byte(" \t\r\n"),
		[]byte("null"),
		[]byte("null true"),
		[]byte("{\"a\":}"),
		{0xff},
		{'"', 0xff, '"'},
	}
	for _, value := range tests {
		if _, err := NewJSONLiteral(value); err == nil {
			t.Fatalf("invalid literal accepted: %q", value)
		}
	}
}

func TestLiteralEnforcesAbsoluteRepresentationBound(t *testing.T) {
	if _, err := NewTextLiteral(strings.Repeat("a", maximumLiteralBytes)); err != nil {
		t.Fatalf("maximum text literal rejected: %v", err)
	}
	if _, err := NewTextLiteral(strings.Repeat("a", maximumLiteralBytes+1)); err == nil {
		t.Fatal("oversized text literal accepted")
	}
	maximumJSON := `"` + strings.Repeat("a", maximumLiteralBytes-2) + `"`
	if _, err := NewJSONLiteral([]byte(maximumJSON)); err != nil {
		t.Fatalf("maximum JSON literal rejected: %v", err)
	}
	if _, err := NewJSONLiteral([]byte(maximumJSON + " ")); err == nil {
		t.Fatal("oversized JSON input accepted")
	}
}

func TestLiteralRejectsDuplicateObjectMembersForInternalCallers(t *testing.T) {
	if _, err := NewJSONLiteral([]byte(`{"a":1,"a":2}`)); err == nil {
		t.Fatal("literal accepted duplicate member outside HTTP preflight")
	}
}

func TestLiteralAccessorsRejectZeroAndTamperedState(t *testing.T) {
	invalid := []Literal{{}, {representation: literalRepresentationText, text: "x"}}
	for _, literal := range invalid {
		if literal.Validate() == nil {
			t.Fatal("invalid literal validated")
		}
		if _, _, err := literal.JSONValue(); err == nil {
			t.Fatal("invalid JSON accessor succeeded")
		}
		if _, _, err := literal.TextValue(); err == nil {
			t.Fatal("invalid text accessor succeeded")
		}
	}
}

func TestLiteralFormattingIsRedacted(t *testing.T) {
	literal, err := NewJSONLiteral([]byte(`"secret"`))
	if err != nil {
		t.Fatal(err)
	}
	for _, formatted := range []string{
		literal.String(), literal.GoString(), fmt.Sprint(literal),
		fmt.Sprintf("%#v", literal), fmt.Sprintf("%q", literal), fmt.Sprintf("%x", literal),
	} {
		if formatted != queryValueRedaction {
			t.Fatalf("literal formatting leaked value: %q", formatted)
		}
	}
}

func TestLiteralCannotBeDirectlyJSONMarshaled(t *testing.T) {
	literal, err := NewJSONLiteral([]byte(`"secret"`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err = json.Marshal(literal); err == nil {
		t.Fatal("opaque literal was silently marshaled")
	} else if strings.Contains(err.Error(), "secret") {
		t.Fatalf("marshal error leaked literal: %v", err)
	}
}
