package offering

import (
	"context"
	"errors"
	"testing"

	offeringapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/offering"
)

func TestCreateRequiresBoundGlobalScope(t *testing.T) {
	storage := &store{allowed: func(context.Context) bool { return false }}
	_, err := storage.Create(context.Background(), offeringapp.CreateMutation{})
	if !errors.Is(err, offeringapp.ErrConfiguration) {
		t.Fatalf("create error = %v", err)
	}
}
