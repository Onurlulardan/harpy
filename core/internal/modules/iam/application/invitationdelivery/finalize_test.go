package invitationdelivery

import (
	"context"
	"errors"
	"testing"
)

func TestFinalizeRejectsNilService(t *testing.T) {
	var service *Service
	if _, err := service.Finalize(context.Background(), FinalizeCommand{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("finalize error = %v", err)
	}
}
