package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestResolveVersionFailsClosedWithoutApplication(t *testing.T) {
	var application *Application
	_, err := application.ResolveVersion(context.Background(), platformtx.TenantScope{}, "package", "version")
	if !errors.Is(err, packageregistry.ErrUnavailable) {
		t.Fatal(err)
	}
}
