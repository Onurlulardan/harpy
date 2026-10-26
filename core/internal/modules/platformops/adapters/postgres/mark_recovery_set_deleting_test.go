package postgres

import (
	"context"
	"testing"
	"time"
)

func TestMarkRecoverySetDeleting(t *testing.T) {
	database := &operationDatabase{row: func(values ...any) error {
		*(values[0].(*bool)) = true
		return nil
	}}
	err := operationStore(database).MarkRecoverySetDeleting(
		context.Background(), "set-id", "manifest", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
}
