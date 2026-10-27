package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

func TestReadPackageBySlugRejectsInvalidSlug(t *testing.T) {
	_, err := (&store{}).ReadPackageBySlug(context.Background(), "INVALID")
	if !errors.Is(err, packageregistry.ErrInvalid) {
		t.Fatalf("read package by slug error = %v", err)
	}
}
