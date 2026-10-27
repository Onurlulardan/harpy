package invitationdelivery

import (
	"context"
	"errors"
	"testing"

	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
)

func TestClaimRejectsInvalidCommand(t *testing.T) {
	var value store
	if _, err := value.Claim(context.Background(), invitationdeliveryapp.ClaimCommand{}); !errors.Is(
		err, invitationdeliveryapp.ErrInvalid,
	) {
		t.Fatalf("claim error = %v", err)
	}
}
