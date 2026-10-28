package strictjson

import (
	"strings"
	"testing"
)

type fixture struct {
	Version int `json:"version"`
}

func TestDecodeAcceptsSingleStrictValue(t *testing.T) {
	var target fixture
	if err := Decode([]byte(`{"version":1}`), &target); err != nil {
		t.Fatal(err)
	}
	if target.Version != 1 {
		t.Fatalf("expected version 1, got %d", target.Version)
	}
}

func TestDecodeRejectsDuplicateKey(t *testing.T) {
	assertDecodeError(t, `{"version":1,"version":2}`, "duplicate key")
}

func TestDecodeRejectsUnknownMember(t *testing.T) {
	assertDecodeError(t, `{"version":1,"extra":true}`, "not an exact schema key")
}

func TestDecodeRejectsCaseFoldedMember(t *testing.T) {
	assertDecodeError(t, `{"Version":1}`, "not an exact schema key")
}

func TestDecodeRejectsTrailingValue(t *testing.T) {
	assertDecodeError(t, `{"version":1} {"version":2}`, "trailing")
}

func assertDecodeError(t *testing.T, input, expected string) {
	t.Helper()
	var target fixture
	err := Decode([]byte(input), &target)
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("expected %q error, got %v", expected, err)
	}
}
