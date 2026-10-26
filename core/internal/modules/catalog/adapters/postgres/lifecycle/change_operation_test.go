package lifecycle

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
)

func TestChangeOperationRejectsEmptyOperation(t *testing.T) {
	_, err := (&store{}).ChangeOperation(context.Background(), catalogapp.OperationMutation{})
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatal(err)
	}
}
