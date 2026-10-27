package postrestore

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestChallengePinsCurrentRestoreBinding(t *testing.T) {
	now := time.Date(2026, 7, 23, 10, 0, 0, 0, time.UTC)
	environmentID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	restoreID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	operationID := uuid.MustParse("33333333-3333-4333-8333-333333333333")
	store := &postRestoreStoreFixture{binding: Binding{
		TenantID:            uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		OwnerUserID:         uuid.MustParse("55555555-5555-4555-8555-555555555555"),
		AnchorID:            uuid.MustParse("66666666-6666-4666-8666-666666666666"),
		NewDatabaseSystemID: uuid.MustParse("77777777-7777-4777-8777-777777777777"),
		SecurityEpoch:       [32]byte{9},
	}}
	service := newPostRestoreServiceFixture(t, store, now, operationID)
	result, err := service.Challenge(t.Context(), environmentID, restoreID,
		"https://console.example.com")
	if err != nil {
		t.Fatal(err)
	}
	command := store.issuedCommand
	want := challengeContextDigest(environmentID, restoreID, store.binding,
		"console.example.com", "https://console.example.com", command.Challenge)
	if command.ContextSHA256 != want || command.OperationID != operationID ||
		!command.ExpiresAt.Equal(now.Add(5*time.Minute)) || result.OperationID != operationID {
		t.Fatalf("challenge was not bound: command=%+v result=%+v", command, result)
	}
}
