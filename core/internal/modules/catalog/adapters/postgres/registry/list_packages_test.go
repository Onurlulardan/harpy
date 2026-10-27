package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

func TestListPackagesRejectsUnboundedPage(t *testing.T) {
	_, err := (&store{}).ListPackages(context.Background(), 0, "")
	if !errors.Is(err, packageregistry.ErrInvalid) {
		t.Fatalf("list packages error = %v", err)
	}
}
