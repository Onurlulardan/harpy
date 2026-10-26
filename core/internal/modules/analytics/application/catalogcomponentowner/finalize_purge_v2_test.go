package catalogcomponentowner

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestFinalizeRejectsInvalidCommand(t *testing.T) {
	owner := &Owner{}
	if _, err := owner.FinalizeComponentPurge(
		context.Background(),
		platformtx.TenantScope{},
		componentowner.PurgeCommand{},
	); err == nil {
		t.Fatal("finalize accepted an invalid command")
	}
}
