package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestListPackagesFailsClosedWithoutApplication(t *testing.T) {
	var application *Application
	_, err := application.ListPackages(context.Background(), platformtx.TenantScope{}, 50, "")
	if !errors.Is(err, packageregistry.ErrUnavailable) {
		t.Fatal(err)
	}
}
