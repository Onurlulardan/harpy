package runtime

import (
	"testing"

	queryapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGlobalExecuteAggregateRejectsMissingEngine(t *testing.T) {
	var engine *GlobalEngine
	_, err := engine.ExecuteAggregate(
		t.Context(), platformtx.GlobalScope{}, "billing.revenue_projection",
		queryapi.QueryRequest{},
	)
	if err == nil {
		t.Fatal("missing global aggregate engine was accepted")
	}
}
