package runtime

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestValidateRejectsEmptyContracts(t *testing.T) {
	if err := Validate(api.QueryRequest{}, api.CatalogSet{}); err == nil {
		t.Fatal("empty REST Query contracts were accepted")
	}
}
