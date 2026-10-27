package resource

import (
	"errors"
	"testing"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestVerifierFailsClosedOutsideOwnedCommandScope(t *testing.T) {
	verifier := NewVerifier()
	version := uuid.MustParse("019b0000-0000-7000-8000-000000000201")
	if _, err := verifier.VerifyImmutable(t.Context(), platformtx.TenantScope{},
		publicationapp.ResourceQuery, version); !errors.Is(err, publicationapp.ErrConflict) {
		t.Fatalf("immutable error = %v", err)
	}
	if _, err := verifier.VerifyPublished(t.Context(), platformtx.TenantScope{},
		[]uuid.UUID{version}); !errors.Is(err, dashboardapp.ErrConflict) {
		t.Fatalf("published error = %v", err)
	}
}
