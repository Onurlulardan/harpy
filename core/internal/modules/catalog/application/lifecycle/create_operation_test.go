package lifecycle

import (
	"context"
	"errors"
	"testing"
)

func TestCreateOperationRejectsMissingStore(t *testing.T) {
	_, err := CreateOperation(context.Background(), nil, StartMutation{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
}
