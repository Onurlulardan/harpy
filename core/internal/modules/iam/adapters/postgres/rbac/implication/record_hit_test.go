package implicationpostgres

import (
	"context"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

func TestRecordRejectsInvalidInput(t *testing.T) {
	var storage *store
	if _, err := storage.RecordHit(context.Background(), tenantScope{}, rbacdomain.RecordHitCommand{}); err == nil {
		t.Fatal("record_hit accepted invalid input")
	}
}
