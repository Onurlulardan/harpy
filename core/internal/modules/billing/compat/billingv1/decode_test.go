package billingv1

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

type bodyFixture struct {
	Value string `json:"value"`
}

func TestDecodeBodyEnforcesExactByteLimit(t *testing.T) {
	t.Parallel()
	const framing = `{"value":""}`
	exact := `{"value":"` +
		strings.Repeat("a", int(maximumBodyBytes)-len(framing)) + `"}`
	request := httptest.NewRequest("POST", "/", strings.NewReader(exact))
	decoded, err := decodeBody[bodyFixture](request)
	if err != nil || len(decoded.Value) != int(maximumBodyBytes)-len(framing) {
		t.Fatalf("exact body decoded length=%d error=%v", len(decoded.Value), err)
	}
	over := exact[:len(exact)-2] + "a" + exact[len(exact)-2:]
	request = httptest.NewRequest("POST", "/", strings.NewReader(over))
	if _, err := decodeBody[bodyFixture](request); !errors.Is(err, ErrRequestInvalid) {
		t.Fatalf("oversized body error = %v", err)
	}
}

func TestDecodeBodyRejectsTrailingDocument(t *testing.T) {
	t.Parallel()
	request := httptest.NewRequest("POST", "/", strings.NewReader(
		`{"value":"one"}{"value":"two"}`,
	))
	if _, err := decodeBody[bodyFixture](request); !errors.Is(err, ErrRequestInvalid) {
		t.Fatalf("trailing document error = %v", err)
	}
}
