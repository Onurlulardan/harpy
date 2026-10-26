package copy

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestStartRequiresBoundStore(t *testing.T) {
	var store *Store
	if _, err := store.Start(context.Background(), ports.RecoveryExecutionIdentity{}); err == nil {
		t.Fatal("unbound recovery copy store accepted")
	}
}
