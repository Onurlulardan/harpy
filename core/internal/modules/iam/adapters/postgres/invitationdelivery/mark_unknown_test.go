package invitationdelivery

import (
	"context"
	"errors"
	"testing"

	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
)

func TestMarkUnknownRejectsInvalidReference(t *testing.T) {
	var value store
	if err := value.MarkUnknown(context.Background(), invitationdeliveryapp.EffectReference{}); !errors.Is(
		err, invitationdeliveryapp.ErrInvalid,
	) {
		t.Fatalf("mark unknown error = %v", err)
	}
}
