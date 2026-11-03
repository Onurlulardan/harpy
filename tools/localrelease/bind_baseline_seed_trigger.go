package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	assumeEnvironmentOwnerSQL    = "SET LOCAL ROLE harpy_owner"
	suspendEnvironmentTriggerSQL = `ALTER TABLE platform.environment_identity
DISABLE TRIGGER environment_identity_reject_mutation`
	resumeEnvironmentTriggerSQL = `ALTER TABLE platform.environment_identity
ENABLE TRIGGER environment_identity_reject_mutation`
)

func suspendEnvironmentIdentityTrigger(ctx context.Context, transaction pgx.Tx) error {
	if _, err := transaction.Exec(ctx, assumeEnvironmentOwnerSQL); err != nil {
		return fmt.Errorf("bind baseline seed identity to the live database: %w", err)
	}
	if _, err := transaction.Exec(ctx, suspendEnvironmentTriggerSQL); err != nil {
		return fmt.Errorf("bind baseline seed identity to the live database: %w", err)
	}
	return nil
}

func resumeEnvironmentIdentityTrigger(ctx context.Context, transaction pgx.Tx) error {
	if _, err := transaction.Exec(ctx, resumeEnvironmentTriggerSQL); err != nil {
		return fmt.Errorf("bind baseline seed identity to the live database: %w", err)
	}
	var enabled bool
	err := transaction.QueryRow(ctx, `SELECT trigger_row.tgenabled='O'
FROM pg_catalog.pg_trigger AS trigger_row
WHERE trigger_row.tgname='environment_identity_reject_mutation'
  AND trigger_row.tgrelid='platform.environment_identity'::regclass
  AND NOT trigger_row.tgisinternal`).Scan(&enabled)
	if err != nil || !enabled {
		return fmt.Errorf("bind baseline seed identity to the live database: immutable trigger is not restored")
	}
	return nil
}
