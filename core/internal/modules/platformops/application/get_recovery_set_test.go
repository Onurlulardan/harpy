package application

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

func TestGetRecoverySet(t *testing.T) {
	id := appUUID(6)
	store := &recoveryControlFixture{sets: map[string]ports.RecoverySetRecord{id: {
		RecoverySetID: id, EnvironmentID: appUUID(1), Sequence: 3,
		State: "protected", ManifestSHA256: appDigest("a"),
	}}, restores: map[string]ports.RestoreRecord{}}
	service := NewRecoveryService(store, store, store, store, recoveryClockFixture{})
	result, err := service.GetRecoverySet(context.Background(), id)
	if err != nil || result.RecoverySetID != id || result.State != "protected" {
		t.Fatalf("get recovery set = %#v %v", result, err)
	}
}
