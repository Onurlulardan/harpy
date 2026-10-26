package billing

import (
	"errors"
	"testing"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/google/uuid"
)

func TestDocumentsUsageStringV7IsStrict(t *testing.T) {
	valid, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	if parsed, parseErr := documentsUsageStringV7(valid.String()); parseErr != nil || parsed != valid {
		t.Fatalf("strict UUIDv7 rejected: id=%s error=%v", parsed, parseErr)
	}
	for _, invalid := range []string{uuid.New().String(), "not-a-uuid",
		valid.String()[:35] + "A"} {
		if _, parseErr := documentsUsageStringV7(invalid); !errors.Is(
			parseErr, usageapp.ErrRequest,
		) {
			t.Fatalf("invalid identifier %q passed: %v", invalid, parseErr)
		}
	}
}

func TestDocumentsUsageDeniedMapsToHardDenial(t *testing.T) {
	allowed, err := documentsUsageAllowed(ErrDocumentsUsageDenied)
	if err != nil || allowed {
		t.Fatalf("denial mapping allowed=%v error=%v", allowed, err)
	}
	probe := errors.New("dependency failed")
	allowed, err = documentsUsageAllowed(probe)
	if allowed || !errors.Is(err, probe) {
		t.Fatalf("dependency mapping allowed=%v error=%v", allowed, err)
	}
}
