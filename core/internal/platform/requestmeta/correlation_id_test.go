package requestmeta

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestResolveCorrelationIDAcceptsExactValidValue(t *testing.T) {
	header := make(http.Header)
	header.Set(CorrelationHeader, "client_123.example")
	resolved, err := ResolveCorrelationID(header)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != "client_123.example" {
		t.Fatalf("unexpected correlation ID: %s", resolved)
	}
}

func TestResolveCorrelationIDReplacesInvalidAndDuplicateValues(t *testing.T) {
	for _, values := range [][]string{
		{"contains space"},
		{"non-ascii-ğ"},
		{strings.Repeat("a", 65)},
		{"first", "second"},
	} {
		header := http.Header{CorrelationHeader: values}
		resolved, err := ResolveCorrelationID(header)
		if err != nil {
			t.Fatal(err)
		}
		if len(resolved) != 36 || resolved[14] != '4' || !strings.ContainsRune("89ab", rune(resolved[19])) {
			t.Fatalf("invalid UUIDv4 replacement: %s", resolved)
		}
	}
}

func TestResolveCorrelationIDReportsEntropyFailure(t *testing.T) {
	_, err := resolveCorrelationID(make(http.Header), func() (string, error) {
		return "", errors.New("entropy unavailable")
	})
	if err == nil {
		t.Fatal("expected entropy failure")
	}
}
