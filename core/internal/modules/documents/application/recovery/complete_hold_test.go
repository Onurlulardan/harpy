package recovery

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestCompleteHoldRequiresExactStoreTransition(t *testing.T) {
	fixture := &storeFixture{}
	service := newFixtureService(fixture)
	setID := uuid.MustParse("30000000-0000-4000-8000-000000000003")
	digest := [32]byte{1}
	if !errors.Is(service.CompleteHold(context.Background(), setID, digest, 0), ErrConflict) {
		t.Fatal("seal must reject an unchanged hold")
	}
	fixture.sealed = true
	if err := service.CompleteHold(context.Background(), setID, digest, 0); err != nil {
		t.Fatalf("complete hold: %v", err)
	}
}
