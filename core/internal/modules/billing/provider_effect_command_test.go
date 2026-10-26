package billing

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/provider/manual"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
)

func TestProviderEffectCommandRunsThroughExternalEffectRegistry(t *testing.T) {
	command, err := NewProviderEffectCommand(
		providerRegistry(t), "manual", manual.CommandOperation,
	)
	if err != nil {
		t.Fatal(err)
	}
	registry, err := effectexecution.NewRegistry([]effectexecution.Provider{command})
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := registry.Resolve(command.Identity())
	if err != nil {
		t.Fatal(err)
	}
	result, err := resolved.Invoke(
		context.Background(), providerEffectRequest(t, manual.CommandOperation),
	)
	if err != nil || result.IsUnknown() {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	receipt, ok := result.Receipt()
	if !ok || receipt.Code != "MANUAL_COMMAND_ACCEPTED" {
		t.Fatalf("receipt=%+v", receipt)
	}
}
