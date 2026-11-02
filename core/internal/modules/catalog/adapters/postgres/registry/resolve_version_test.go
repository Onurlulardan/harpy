package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

func TestResolveVersionRejectsInvalidVersion(t *testing.T) {
	_, err := (&store{}).ResolveVersion(context.Background(),
		"018f5000-0000-7000-8000-000000000001", "latest")
	if !errors.Is(err, packageregistry.ErrInvalid) {
		t.Fatalf("resolve version error = %v", err)
	}
}
