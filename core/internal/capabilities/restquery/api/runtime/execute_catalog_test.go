package runtime

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestExecuteCatalogRejectsMissingEngine(t *testing.T) {
	var engine *Engine
	if _, err := engine.ExecuteCatalog(
		t.Context(), platformtx.TenantScope{}, api.CatalogSet{}, api.QueryRequest{},
	); err == nil {
		t.Fatal("missing exact engine was accepted")
	}
}
