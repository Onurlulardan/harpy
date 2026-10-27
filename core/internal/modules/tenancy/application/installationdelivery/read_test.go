package installationdelivery

import (
	"context"
	"errors"
	"testing"
)

func TestReadRejectsInvalidQuery(t *testing.T) {
	if _, err := Read(context.Background(), nil, ReadQuery{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Read() error = %v, want %v", err, ErrInvalid)
	}
}
