package snapshot

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

type storeFixture bool

func (fixture storeFixture) Start(
	context.Context, ports.RecoveryExecutionIdentity,
) (bool, error) {
	return bool(fixture), nil
}

func TestStartRequiresExactAttemptLeaseTransition(t *testing.T) {
	service := New(storeFixture(true))
	err := service.Start(context.Background(), ports.RecoveryExecutionIdentity{
		RecoverySetID: "10000000-0000-4000-8000-000000000001",
		AttemptID:     "01890f3e-7b10-7cc0-98c4-dc0c0c07398f",
		LeaseToken:    "01890f3e-7b10-7cc0-a8c4-dc0c0c073990",
	})
	if err != nil {
		t.Fatal(err)
	}
}
