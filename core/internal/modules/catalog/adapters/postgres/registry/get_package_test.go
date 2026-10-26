package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

func TestGetRejectsInvalidPackage(t *testing.T) {
	_, err := (&store{}).GetPackage(context.Background(), "not-a-package")
	if !errors.Is(err, packageregistry.ErrInvalid) {
		t.Fatalf("get package error = %v", err)
	}
}
