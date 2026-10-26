package provider

import (
	"context"
	"errors"
	"testing"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
)

func TestCreateCommandRequiresBoundActorTenantStore(t *testing.T) {
	_, err := (&commandStore{}).CreateCommand(
		context.Background(), providerapp.CommandMutation{},
	)
	if !errors.Is(err, providerapp.ErrCommandRequest) {
		t.Fatalf("CreateCommand() error = %v", err)
	}
}
