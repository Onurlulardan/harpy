package catalogcomponentowner

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRestoreRejectsInvalidCommand(t *testing.T) {
	owner := &Owner{}
	if _, err := owner.RestoreComponent(
		context.Background(),
		platformtx.TenantScope{},
		componentowner.LifecycleCommand{},
	); err == nil {
		t.Fatal("restore accepted an invalid command")
	}
}
