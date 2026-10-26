package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestGetPackageFailsClosedWithoutApplication(t *testing.T) {
	var application *Application
	_, err := application.GetPackage(context.Background(), platformtx.TenantScope{}, "package")
	if !errors.Is(err, packageregistry.ErrUnavailable) {
		t.Fatal(err)
	}
}
