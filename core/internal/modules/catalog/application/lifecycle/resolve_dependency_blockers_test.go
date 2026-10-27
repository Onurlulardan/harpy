package lifecycle

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func TestResolveDependencyBlockersRejectsMissingStore(t *testing.T) {
	_, err := ResolveDependencyBlockers(context.Background(), nil, packagecontract.PackageID{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
}
