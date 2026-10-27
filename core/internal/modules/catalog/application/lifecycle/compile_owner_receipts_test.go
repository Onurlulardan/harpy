package lifecycle

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestCompileOwnerReceiptsRejectsMissingStore(t *testing.T) {
	_, err := CompileOwnerReceipts(context.Background(), nil, uuid.Nil)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
}
