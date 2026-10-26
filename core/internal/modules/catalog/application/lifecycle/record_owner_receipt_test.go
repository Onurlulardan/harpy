package lifecycle

import (
	"context"
	"errors"
	"testing"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
)

func TestRecordOwnerReceiptRejectsMissingStore(t *testing.T) {
	err := RecordOwnerReceipt(context.Background(), nil, catalogdomain.OwnerReceipt{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
}
