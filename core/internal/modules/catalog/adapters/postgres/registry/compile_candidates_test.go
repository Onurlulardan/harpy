package registry

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

func TestCompileCandidatesRejectsUnboundedRequest(t *testing.T) {
	_, err := (&store{}).CompileCandidates(context.Background(),
		"018f5000-0000-7000-8000-000000000001", 65)
	if !errors.Is(err, packageregistry.ErrInvalid) {
		t.Fatalf("compile candidates error = %v", err)
	}
}
