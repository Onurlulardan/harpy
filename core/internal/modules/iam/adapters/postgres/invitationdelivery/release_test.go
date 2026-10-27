package invitationdelivery

import (
	"context"
	"errors"
	"testing"

	invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"
)

func TestReleaseRejectsInvalidReference(t *testing.T) {
	var value store
	if err := value.Release(context.Background(), invitationdeliveryapp.EffectReference{}); !errors.Is(
		err, invitationdeliveryapp.ErrInvalid,
	) {
		t.Fatalf("release error = %v", err)
	}
}
