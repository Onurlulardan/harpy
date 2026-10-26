package postgres

import (
	"context"
	"errors"
	"testing"
)

func TestCreateVersionTableRequiresCallerTransaction(t *testing.T) {
	store, err := NewStore(testRunID, validSelectedMigrations())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.CreateVersionTable(context.Background(), nil); !errors.Is(
		err, ErrCreateTransactionRequired,
	) {
		t.Fatalf("non-transactional history create accepted: %v", err)
	}
}
