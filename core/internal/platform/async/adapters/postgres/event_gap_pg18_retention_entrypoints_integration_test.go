//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func assertRetiredGapRedactionEntrypoints(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	jobID uuid.UUID,
	wantState string,
) {
	t.Helper()
	var state string
	if err := fixture.owner.QueryRowContext(
		ctx, "SELECT state FROM platform.jobs WHERE id=$1", jobID,
	).Scan(&state); err != nil || state != wantState {
		t.Fatalf("gap retention state = %s, want %s: %v", state, wantState, err)
	}
	for _, routine := range []string{
		"platform.platform_async_redact_outbox_payloads_v1",
		"platform.platform_async_redact_dead_letter_payloads_v1",
	} {
		var ignored int
		err := fixture.runtime.QueryRow(
			ctx, "SELECT "+routine+"(500)",
		).Scan(&ignored)
		var databaseError *pgconn.PgError
		if !errors.As(err, &databaseError) || databaseError.Code != "42501" {
			t.Fatalf("retired retention entrypoint %s = %v", routine, err)
		}
	}
}
