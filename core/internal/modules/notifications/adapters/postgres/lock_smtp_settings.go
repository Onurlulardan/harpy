package postgres

import (
	"context"

	"github.com/google/uuid"
)

func (store *Store) LockSMTPSettings(ctx context.Context) error {
	if !store.usable(ctx, true) || store.tenantID == uuid.Nil {
		return ErrInvalid
	}
	var singleton bool
	var locked bool
	if err := store.queryRow(ctx, lockSMTPSettingsSQL,
		store.tenantID.String()).Scan(&singleton, &locked); err != nil {
		return err
	}
	if !singleton || !locked {
		return ErrConflict
	}
	return nil
}

func (store *Store) LatestSMTPSettingsRevision(ctx context.Context) (int64, error) {
	if !store.usable(ctx, false) {
		return 0, ErrInvalid
	}
	var revision int64
	if err := store.queryRow(ctx, latestSMTPSettingsRevisionSQL).Scan(&revision); err != nil ||
		revision < 0 {
		return 0, ErrInvalid
	}
	return revision, nil
}
