package lifecycle

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
)

func TestApprovePurgeRejectsEmptyReceipt(t *testing.T) {
	_, _, err := (&store{}).ApprovePurge(
		context.Background(), catalogapp.PurgeApprovalMutation{},
	)
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatalf("empty approval error = %v", err)
	}
}
