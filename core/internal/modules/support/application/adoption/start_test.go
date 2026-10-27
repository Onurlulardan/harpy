package adoption

import (
	"context"
	"errors"
	"testing"
)

func TestStartRejectsUnavailableWorker(t *testing.T) {
	_, _, err := (*Worker)(nil).Start(context.Background(), StartRequest{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("start error = %v, want invalid", err)
	}
}
