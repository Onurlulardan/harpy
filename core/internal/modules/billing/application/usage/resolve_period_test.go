package usage

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

func TestResolvePeriodReturnsOnlyStoreUUIDv7(t *testing.T) {
	commands, store, _, _ := reserveFixture(t, entitlement.Hard, 0, 0, 10)
	key, _ := entitlement.NewKey("forms.submissions")
	periodID, err := commands.ResolvePeriod(context.Background(), key)
	if err != nil || periodID != store.snapshot.PeriodID || store.resolveCalls != 1 {
		t.Fatalf("period=%s calls=%d error=%v", periodID, store.resolveCalls, err)
	}
}
