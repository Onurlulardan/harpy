package discountcatalog

import (
	"context"
	"errors"
	"testing"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
)

func TestRetireRequiresBoundGlobalScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, err := storage.Retire(context.Background(), discountapp.RetireMutation{})
	if !errors.Is(err, discountapp.ErrConfiguration) {
		t.Fatalf("retire error = %v", err)
	}
}
