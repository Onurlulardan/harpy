package customerprovisioning

import (
	"context"
	"errors"
	"testing"
)

func TestCreateRejectsInvalidCommand(t *testing.T) {
	var coordinator *Coordinator
	if _, err := coordinator.Create(context.Background(), Command{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalid)
	}
}
