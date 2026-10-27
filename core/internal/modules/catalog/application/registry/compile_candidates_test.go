package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCompileCandidatesFailsClosedWithoutApplication(t *testing.T) {
	var application *Application
	_, err := application.CompileCandidates(context.Background(), platformtx.TenantScope{}, "package", 64)
	if !errors.Is(err, packageregistry.ErrUnavailable) {
		t.Fatal(err)
	}
}
