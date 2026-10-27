package invitationdelivery

import (
	"context"
	"errors"
	"testing"
)

func TestReleaseRejectsNilService(t *testing.T) {
	var service *Service
	if err := service.Release(context.Background(), EffectReference{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("release error = %v", err)
	}
}
