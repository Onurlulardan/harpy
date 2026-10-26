package lifecycle

import (
	"context"
	"errors"
	"testing"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
)

func TestGetRejectsEmptyPackage(t *testing.T) {
	_, err := (&store{}).GetInstallation(context.Background(), packagecontract.PackageID{})
	if !errors.Is(err, catalogapp.ErrInvalid) {
		t.Fatalf("get installation error = %v", err)
	}
}
