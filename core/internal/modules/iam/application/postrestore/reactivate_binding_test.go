package postrestore

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestReactivateRejectsStoredChallengeBindingMismatch(t *testing.T) {
	now := time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)
	environmentID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	restoreID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	operationID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	binding := Binding{
		TenantID:            uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		OwnerUserID:         uuid.MustParse("55555555-5555-4555-8555-555555555555"),
		AnchorID:            uuid.MustParse("66666666-6666-4666-8666-666666666666"),
		NewDatabaseSystemID: uuid.MustParse("77777777-7777-4777-8777-777777777777"),
		SecurityEpoch:       [32]byte{9},
	}
	store := &postRestoreStoreFixture{binding: binding,
		context: ChallengeContext{ChallengeRecord: ChallengeRecord{
			OperationID: operationID, TenantID: binding.TenantID,
			OwnerUserID: binding.OwnerUserID, AnchorID: binding.AnchorID,
			Challenge: "stored-challenge", RPID: "console.example.com",
			ContextSHA256: [32]byte{1},
		}}}
	service := newPostRestoreServiceFixture(t, store, now, operationID)
	_, err := service.Reactivate(t.Context(), environmentID, restoreID, operationID,
		"https://console.example.com", []byte(`{}`), []byte("a-long-new-password"))
	if !errors.Is(err, ErrDenied) || store.failures != 0 {
		t.Fatalf("error=%v failures=%d", err, store.failures)
	}
}
