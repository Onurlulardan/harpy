package foundation

import (
	"context"
	"testing"
)

func TestExternalEffectsAllowedRequiresNetworkScope(t *testing.T) {
	if !ExternalEffectsAllowed(context.Background()) || ExternalEffectsAllowed(nil) ||
		DatabaseEffectsAllowed(context.Background()) || DatabaseEffectsAllowed(nil) {
		t.Fatal("Documents external-effect context policy mismatch")
	}
}
