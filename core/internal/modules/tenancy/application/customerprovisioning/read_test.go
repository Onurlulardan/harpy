package customerprovisioning

import (
	"context"
	"errors"
	"testing"
)

func TestReadRejectsInvalidQuery(t *testing.T) {
	var coordinator *Coordinator
	if _, err := coordinator.Read(context.Background(), ReadQuery{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Read() error = %v, want %v", err, ErrInvalid)
	}
}
