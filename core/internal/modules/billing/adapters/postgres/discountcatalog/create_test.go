package discountcatalog

import (
	"context"
	"errors"
	"testing"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
)

func TestCreateRequiresBoundGlobalScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, err := storage.Create(context.Background(), discountapp.CreateMutation{})
	if !errors.Is(err, discountapp.ErrConfiguration) {
		t.Fatalf("create error = %v", err)
	}
}
