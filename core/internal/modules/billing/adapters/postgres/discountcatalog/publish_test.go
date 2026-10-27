package discountcatalog

import (
	"context"
	"errors"
	"testing"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discountcatalog"
)

func TestPublishRequiresBoundGlobalScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, err := storage.Publish(context.Background(), discountapp.CreateVersionMutation{})
	if !errors.Is(err, discountapp.ErrConfiguration) {
		t.Fatalf("publish error = %v", err)
	}
}
