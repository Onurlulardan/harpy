package randomid

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestNewUUIDv4SetsVersionAndVariant(t *testing.T) {
	resolved, err := newUUIDv4(bytes.NewReader(make([]byte, 16)))
	if err != nil {
		t.Fatal(err)
	}
	if resolved != "00000000-0000-4000-8000-000000000000" {
		t.Fatalf("unexpected deterministic UUID: %s", resolved)
	}
}

func TestNewUUIDv4ReportsEntropyFailure(t *testing.T) {
	reader := io.MultiReader(bytes.NewReader(make([]byte, 4)), errorReader{})
	if _, err := newUUIDv4(reader); err == nil {
		t.Fatal("expected entropy failure")
	}
}

func TestNewUUIDv4DoesNotRepeatInSample(t *testing.T) {
	seen := make(map[string]bool, 1024)
	for index := 0; index < 1024; index++ {
		value, err := NewUUIDv4()
		if err != nil {
			t.Fatal(err)
		}
		if seen[value] {
			t.Fatalf("duplicate UUID: %s", value)
		}
		seen[value] = true
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errors.New("entropy unavailable")
}
