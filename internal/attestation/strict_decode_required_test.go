package attestation

import (
	"encoding/json"
	"testing"
)

type requiredZeroFixture struct {
	Enabled bool `json:"enabled"`
	Count   int  `json:"count"`
}

type explicitNullableFixture struct {
	Name *string `json:"name" harpyNull:"allowed"`
}

type implicitNullableFixture struct {
	Name *string `json:"name"`
}

type explicitRawNullableFixture struct {
	Value json.RawMessage `json:"value" harpyNull:"allowed"`
}

type invalidScalarNullableFixture struct {
	Value string `json:"value" harpyNull:"allowed"`
}

func TestDecodeStrictRequiresExplicitZeroValueMembers(t *testing.T) {
	var decoded requiredZeroFixture
	if err := DecodeStrict([]byte(`{"enabled":false,"count":0}`), &decoded); err != nil {
		t.Fatalf("explicit zero values rejected: %v", err)
	}
	for _, document := range []string{
		`{"enabled":false}`,
		`{"count":0}`,
		`{"Enabled":false,"count":0}`,
		`{"enabled":null,"count":0}`,
	} {
		if err := DecodeStrict([]byte(document), &decoded); err == nil {
			t.Fatalf("non-exact document accepted: %s", document)
		}
	}
}

func TestDecodeStrictDistinguishesExplicitRawNull(t *testing.T) {
	var decoded explicitRawNullableFixture
	if err := DecodeStrict([]byte(`{"value":null}`), &decoded); err != nil {
		t.Fatalf("explicit raw null rejected: %v", err)
	}
	if string(decoded.Value) != "null" {
		t.Fatalf("explicit raw null decoded as %q", decoded.Value)
	}
	if err := DecodeStrict([]byte(`{}`), &decoded); err == nil {
		t.Fatal("missing raw nullable member accepted")
	}
	var scalar invalidScalarNullableFixture
	if err := DecodeStrict([]byte(`{"value":null}`), &scalar); err == nil {
		t.Fatal("null accepted for tagged non-pointer scalar")
	}
}

func TestDecodeStrictAllowsOnlyExplicitNullableMembers(t *testing.T) {
	var decoded explicitNullableFixture
	if err := DecodeStrict([]byte(`{"name":null}`), &decoded); err != nil {
		t.Fatalf("explicit nullable member rejected: %v", err)
	}
	if decoded.Name != nil {
		t.Fatal("null member did not decode to nil")
	}
	var nilTarget *explicitNullableFixture
	if err := DecodeStrict([]byte(`{"name":null}`), nilTarget); err == nil {
		t.Fatal("nil decode target accepted")
	}
	if err := DecodeStrict([]byte(`null`), &decoded); err == nil {
		t.Fatal("null top-level document accepted")
	}
	var implicit implicitNullableFixture
	if err := DecodeStrict([]byte(`{"name":null}`), &implicit); err == nil {
		t.Fatal("untagged nullable member accepted")
	}
}
