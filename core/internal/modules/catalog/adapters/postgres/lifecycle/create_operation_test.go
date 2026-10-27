package lifecycle

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
)

func TestCreateRejectsEmptyOperation(t *testing.T) {
	_, err := (&store{}).CreateOperation(context.Background(), catalogapp.StartMutation{})
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatalf("create operation error = %v", err)
	}
}
