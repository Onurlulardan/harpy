package postgres

import (
	"context"
	"errors"
	"testing"
)

func TestDeleteAlwaysReturnsTypedDownForbiddenError(t *testing.T) {
	store, err := NewStore(testRunID, validSelectedMigrations())
	if err != nil {
		t.Fatal(err)
	}
	err = store.Delete(context.Background(), nil, 20260714000100)
	if !errors.Is(err, ErrDownForbidden) {
		t.Fatalf("delete did not return down-forbidden error: %v", err)
	}
	var typed *DownForbiddenError
	if !errors.As(err, &typed) || typed.Version != 20260714000100 {
		t.Fatalf("delete error lost migration version: %#v", typed)
	}
}
