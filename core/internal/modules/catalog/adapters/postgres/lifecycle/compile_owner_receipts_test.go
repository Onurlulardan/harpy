package lifecycle

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	"github.com/google/uuid"
)

func TestCompileOwnerReceiptsRejectsEmptyOperation(t *testing.T) {
	_, err := (&store{}).CompileOwnerReceipts(context.Background(), uuid.Nil)
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatal(err)
	}
}
