package invitationdelivery

import (
	"context"
	"errors"
	"testing"
)

func TestMarkUnknownRejectsNilService(t *testing.T) {
	var service *Service
	if err := service.MarkUnknown(context.Background(), EffectReference{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("mark unknown error = %v", err)
	}
}
