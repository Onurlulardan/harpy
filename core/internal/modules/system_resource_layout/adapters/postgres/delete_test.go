package postgres

import (
	"context"
	"testing"
)

func TestDeleteRequiresPositiveRoutineOutcome(t *testing.T) {
	store := fixtureLayoutStore(fixtureRow{boolean: false})
	if err := store.Delete(context.Background(), adapterRevisionID, 1); err == nil {
		t.Fatal("false delete outcome was accepted")
	}
	store = fixtureLayoutStore(fixtureRow{boolean: true})
	if err := store.Delete(context.Background(), adapterRevisionID, 1); err != nil {
		t.Fatalf("true delete outcome failed: %v", err)
	}
}
