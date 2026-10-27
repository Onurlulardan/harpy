package systemresourcequery

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
)

func TestValidateCatalogRequestRejectsEmptyContracts(t *testing.T) {
	if err := ValidateCatalogRequest(catalog.Published{}, Request{}); !errors.Is(err, ErrContract) {
		t.Fatalf("ValidateCatalogRequest() error=%v", err)
	}
}
