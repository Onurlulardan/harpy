package tabularmapping

import (
	"errors"
	"slices"
	"testing"
)

func TestProgramAppliesAllDeterministicModes(t *testing.T) {
	document := []byte(`{"contract":"tabular-mapping.v2","fields":[` +
		`{"expression":{"kind":"direct","source":"Name"},"required":true,"target":"name"},` +
		`{"expression":{"kind":"concat","separator":" ","sources":["First","Last"]},"required":true,"target":"full_name"},` +
		`{"expression":{"kind":"static","value":{"code":"NEW"}},"target":"status"},` +
		`{"expression":{"ast":{"arguments":[{"kind":"source","source":"Qty"},{"kind":"literal","value":"2.5"}],"kind":"multiply"},"kind":"formula_numeric"},"required":true,"target":"total"}]}`)
	program, err := Compile(document)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if !slices.Equal(program.Targets(), []string{"name", "full_name", "status", "total"}) {
		t.Fatalf("targets = %#v", program.Targets())
	}
	result, err := program.Apply(
		[]byte(`{"First":"Ada","Last":"Lovelace","Name":"A","Qty":"4"}`),
	)
	want := `{"full_name":"Ada Lovelace","name":"A","status":{"code":"NEW"},"total":"10"}`
	if err != nil || string(result) != want {
		t.Fatalf("mapped = %s, %v", result, err)
	}
}

func TestProgramSupportsOpaqueStableTargets(t *testing.T) {
	program, err := Compile([]byte(`{"contract":"tabular-mapping.v2","fields":[` +
		`{"target":"019b22de-8a00-7000-8000-000000000001",` +
		`"expression":{"kind":"direct","source":"email"}}]}`))
	if err != nil || program.Targets()[0] != "019b22de-8a00-7000-8000-000000000001" {
		t.Fatalf("stable target = %#v, %v", program.Targets(), err)
	}
}

func TestProgramRejectsUnknownAndUnsafeExpressions(t *testing.T) {
	invalid := []string{
		`{"contract":"tabular-mapping.v2","fields":[{"target":"x","expression":{"kind":"script"}}]}`,
		`{"contract":"tabular-mapping.v2","fields":[{"target":"x","extra":true,"expression":{"kind":"static","value":1}}]}`,
		`{"contract":"tabular-mapping.v2","fields":[{"target":"x","expression":{"kind":"formula_numeric","ast":{"kind":"divide","arguments":[{"kind":"literal","value":"1"},{"kind":"literal","value":"0"}]}}}]}`,
	}
	for index, document := range invalid[:2] {
		if _, err := Compile([]byte(document)); !errors.Is(err, ErrInvalid) {
			t.Fatalf("unsafe %d error = %v", index, err)
		}
	}
	program, err := Compile([]byte(invalid[2]))
	if err != nil {
		t.Fatalf("compile row-failing expression: %v", err)
	}
	if _, err := program.Apply([]byte(`{"unused":true}`)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("division error = %v", err)
	}
}

func TestProgramRejectsMissingRequiredSource(t *testing.T) {
	program, err := Compile([]byte(`{"contract":"tabular-mapping.v2","fields":[` +
		`{"target":"x","required":true,"expression":{"kind":"direct","source":"X"}}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := program.Apply([]byte(`{"Y":1}`)); !errors.Is(err, ErrRequiredMissing) {
		t.Fatalf("required error = %v", err)
	}
}
