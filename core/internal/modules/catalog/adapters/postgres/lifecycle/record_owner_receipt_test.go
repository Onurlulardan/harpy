package lifecycle

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
)

func TestRecordRejectsEmptyOwnerReceipt(t *testing.T) {
	err := (&store{}).RecordOwnerReceipt(context.Background(), catalogdomain.OwnerReceipt{})
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatalf("record owner receipt error = %v", err)
	}
}
