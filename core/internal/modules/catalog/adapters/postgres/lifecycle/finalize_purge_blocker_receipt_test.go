package lifecycle

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
)

func TestFinalizePurgeBlockerReceiptRejectsEmptyReceipt(t *testing.T) {
	err := (&store{}).FinalizePurgeBlockerReceipt(context.Background(), catalogapp.PurgeBlockerReceipt{})
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatal(err)
	}
}
