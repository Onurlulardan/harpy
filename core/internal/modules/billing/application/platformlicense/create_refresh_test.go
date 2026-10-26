package platformlicense

import (
	"errors"
	"testing"
	"time"
)

func TestCreateReturnsStableRefreshOperation(t *testing.T) {
	operationID := testV7(t)
	store := &storeFixture{refresh: RefreshAccepted{
		OperationID: operationID,
		Replayed:    true,
		CreatedAt:   time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}}
	service, err := NewService(store)
	if err != nil {
		t.Fatal(err)
	}
	command := RefreshCommand{
		EnvironmentID: testV7(t), TenantID: testV7(t), ActorID: testV7(t),
		IdempotencyKey: "license-refresh-0001",
		RequestSHA256:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	value, err := service.Create(t.Context(), command)
	if err != nil || value.OperationID != operationID ||
		store.refreshInput != command {
		t.Fatalf("refresh=%+v input=%+v err=%v", value, store.refreshInput, err)
	}
}

func TestCreateRejectsInvalidIdempotencyKey(t *testing.T) {
	service, err := NewService(&storeFixture{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Create(t.Context(), RefreshCommand{
		EnvironmentID: testV7(t), TenantID: testV7(t), ActorID: testV7(t),
		IdempotencyKey: "short",
		RequestSHA256:  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("error=%v", err)
	}
}
