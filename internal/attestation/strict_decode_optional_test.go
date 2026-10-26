package attestation

import "testing"

type optionalContract struct {
	Required string  `json:"required"`
	Optional *string `json:"optional" harpyOptional:"allowed"`
}

func TestDecodeStrictAllowsAbsentTaggedOptionalButRejectsNull(t *testing.T) {
	var value optionalContract
	if err := DecodeStrict([]byte(`{"required":"value"}`), &value); err != nil {
		t.Fatalf("absent optional member rejected: %v", err)
	}
	if value.Optional != nil {
		t.Fatal("absent optional member was populated")
	}
	if err := DecodeStrict([]byte(`{"required":"value","optional":null}`), &value); err == nil {
		t.Fatal("null optional member accepted")
	}
}
