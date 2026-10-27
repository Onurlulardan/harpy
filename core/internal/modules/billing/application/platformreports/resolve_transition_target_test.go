package platformreport

import (
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestResolveTransitionTargetRejectsUnconfiguredResolver(t *testing.T) {
	_, err := (&Resolver{}).ResolveTransitionTarget(
		t.Context(), platformtx.GlobalScope{}, platformtx.OwnerAuthority{}, uuid.New(),
	)
	if !errors.Is(err, ErrConfiguration) {
		t.Fatalf("ResolveTransitionTarget error = %v, want ErrConfiguration", err)
	}
}
