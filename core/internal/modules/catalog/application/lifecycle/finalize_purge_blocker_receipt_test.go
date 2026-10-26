package lifecycle

import (
	"context"
	"errors"
	"testing"
)

func TestFinalizePurgeBlockerReceiptRejectsMissingStore(t *testing.T) {
	err := FinalizePurgeBlockerReceipt(context.Background(), nil, PurgeBlockerReceipt{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
}
