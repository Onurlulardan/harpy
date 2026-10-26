package adoption

import (
	"context"
	"errors"
	"testing"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

func TestClaimFailsClosedWithoutAdopterStore(t *testing.T) {
	_, err := (*Store)(nil).AdoptionClaim(context.Background(), application.ClaimCommand{})
	if !errors.Is(err, application.ErrUnavailable) {
		t.Fatalf("claim error = %v, want unavailable", err)
	}
}
