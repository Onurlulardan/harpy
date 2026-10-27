package recovery

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestReleaseHoldRequiresExactStoreTransition(t *testing.T) {
	fixture := &storeFixture{}
	service := newFixtureService(fixture)
	setID := uuid.MustParse("30000000-0000-4000-8000-000000000003")
	digest := [32]byte{1}
	if !errors.Is(service.ReleaseHold(context.Background(), setID, digest), ErrConflict) {
		t.Fatal("unchanged release must conflict")
	}
	fixture.released = true
	if err := service.ReleaseHold(context.Background(), setID, digest); err != nil {
		t.Fatalf("release hold: %v", err)
	}
}
