package provider

import (
	"context"
	"errors"
	"testing"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
)

func TestFinalizeCommandRequiresBoundJobStore(t *testing.T) {
	_, err := (&finalizerStore{}).FinalizeCommand(
		context.Background(), providerapp.FinalizeCommandMutation{},
	)
	if !errors.Is(err, providerapp.ErrCommandRequest) {
		t.Fatalf("FinalizeCommand() error = %v", err)
	}
}
