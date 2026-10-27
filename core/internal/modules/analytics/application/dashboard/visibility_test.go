package dashboard

import (
	"errors"
	"testing"
)

func TestParseVisibilityAcceptsOwnedValues(t *testing.T) {
	for _, value := range []string{"private", "tenant"} {
		visibility, err := ParseVisibility(value)
		if err != nil || string(visibility) != value {
			t.Fatalf("visibility = %q, err = %v", visibility, err)
		}
	}
}

func TestParseVisibilityRejectsLegacyPublicValue(t *testing.T) {
	if _, err := ParseVisibility("public"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("visibility error = %v", err)
	}
}
