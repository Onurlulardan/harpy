package invitationdelivery

import (
	"context"
	"errors"
	"testing"
)

func TestClaimRejectsNilService(t *testing.T) {
	var service *Service
	if _, err := service.Claim(context.Background(), ClaimCommand{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("claim error = %v", err)
	}
}
