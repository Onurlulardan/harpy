package catalogcomponentowner

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRetireRejectsInvalidCommand(t *testing.T) {
	owner := &Owner{}
	if _, err := owner.RetireComponent(
		context.Background(),
		platformtx.TenantScope{},
		componentowner.LifecycleCommand{},
	); err == nil {
		t.Fatal("retire accepted an invalid command")
	}
}
