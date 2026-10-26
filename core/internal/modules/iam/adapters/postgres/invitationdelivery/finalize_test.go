package invitationdelivery

import (
	"context"
	"errors"
	"testing"

	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
)

func TestFinalizeRejectsInvalidCommand(t *testing.T) {
	var value store
	if _, err := value.Finalize(context.Background(), invitationdeliveryapp.FinalizeCommand{}); !errors.Is(
		err, invitationdeliveryapp.ErrInvalid,
	) {
		t.Fatalf("finalize error = %v", err)
	}
}
