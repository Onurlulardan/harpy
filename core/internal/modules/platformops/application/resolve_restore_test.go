package application

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestResolveRestore(t *testing.T) {
	id := appUUID(8)
	store := &recoveryControlFixture{sets: map[string]ports.RecoverySetRecord{},
		restores: map[string]ports.RestoreRecord{id: {
			OperationID: id, RecoverySetID: appUUID(6), EnvironmentID: appUUID(1),
			Disposition: recoveryset.TargetEmpty, State: "verifying",
		}}}
	service := NewRecoveryService(store, store, store, store, recoveryClockFixture{})
	result, err := service.ResolveRestore(context.Background(), id)
	if err != nil || result.OperationID != id || result.State != "verifying" {
		t.Fatalf("get restore = %#v %v", result, err)
	}
}
