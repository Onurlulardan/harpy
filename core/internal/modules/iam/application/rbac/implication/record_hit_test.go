package implication

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRecordRejectsInvalidInput(t *testing.T) {
	var commands *Commands
	if _, err := commands.RecordHit(context.Background(), platformtx.TenantScope{}, rbacdomain.RecordHitCommand{}); err == nil {
		t.Fatal("record_hit accepted invalid input")
	}
}
