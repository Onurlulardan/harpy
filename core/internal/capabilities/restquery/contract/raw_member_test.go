package contract

import (
	"bytes"
	"strings"
	"testing"
)

func TestRawMemberZeroIsAbsent(t *testing.T) {
	t.Parallel()
	var member RawMember
	value, present := member.Value()
	if present || value != nil {
		t.Fatalf("zero member is present: %q", value)
	}
	if err := member.Validate(); err != nil {
		t.Fatalf("absent member is invalid: %v", err)
	}
}

func TestNewRawMemberAcceptsOneExactJSONValue(t *testing.T) {
	t.Parallel()
	values := [][]byte{
		[]byte("null"), []byte("true"), []byte("123"), []byte(`"text"`),
		[]byte(`[1,2]`), []byte(`{"key":"value"}`), []byte("  null \r\n"),
	}
	for _, input := range values {
		member, err := NewRawMember(input)
		if err != nil {
			t.Fatalf("valid JSON %q rejected: %v", input, err)
		}
		actual, present := member.Value()
		if !present || !bytes.Equal(actual, input) {
			t.Fatalf("exact JSON bytes not retained: %q", actual)
		}
	}
}

func TestRawMemberDefensivelyCopiesInputAndOutput(t *testing.T) {
	t.Parallel()
	input := []byte(`{"secret":"stable"}`)
	want := cloneBytes(input)
	member, err := NewRawMember(input)
	if err != nil {
		t.Fatal(err)
	}
	input[2] = 'X'
	actual, _ := member.Value()
	if !bytes.Equal(actual, want) {
		t.Fatalf("constructor retained caller storage: %q", actual)
	}
	actual[2] = 'Y'
	again, _ := member.Value()
	if !bytes.Equal(again, want) {
		t.Fatalf("Value exposed member storage: %q", again)
	}
}

func TestNewRawMemberRejectsInvalidDocumentsWithoutEcho(t *testing.T) {
	t.Parallel()
	values := [][]byte{
		nil, {}, []byte(" \r\n"), []byte(`{"a":`), []byte(`null true`),
		[]byte{0xff}, []byte(`{"password":"super-secret"`),
	}
	for _, input := range values {
		if _, err := NewRawMember(input); err == nil {
			t.Fatalf("invalid raw member %q accepted", input)
		} else if strings.Contains(err.Error(), "super-secret") {
			t.Fatalf("raw member error exposed input: %v", err)
		}
	}
}
