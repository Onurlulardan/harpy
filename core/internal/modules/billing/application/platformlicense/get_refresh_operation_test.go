package platformlicense

import (
	"errors"
	"testing"
	"time"
)

func TestGetReturnsValidatedRefreshState(t *testing.T) {
	environmentID, operationID := testV7(t), testV7(t)
	store := &storeFixture{operation: RefreshOperation{
		OperationID: operationID, State: "pending",
		CreatedAt: time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}}
	service, err := NewService(store)
	if err != nil {
		t.Fatal(err)
	}
	value, err := service.Get(
		t.Context(), environmentID, operationID,
	)
	if err != nil || value.OperationID != operationID ||
		store.environmentID != environmentID || store.operationID != operationID {
		t.Fatalf("operation=%+v err=%v", value, err)
	}
}

func TestGetRejectsUnknownRefreshState(t *testing.T) {
	operationID := testV7(t)
	store := &storeFixture{operation: RefreshOperation{
		OperationID: operationID, State: "invented",
		CreatedAt: time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}}
	service, err := NewService(store)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Get(
		t.Context(), testV7(t), operationID,
	); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("error=%v", err)
	}
}
