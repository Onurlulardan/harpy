package postgres

import (
	"context"

	"github.com/google/uuid"
)

func (transaction *RuntimeTransaction) LockCutoverGroup(
	ctx context.Context,
	tenantID uuid.UUID,
	groupID string,
	expectedGeneration int64,
) error {
	if !transaction.usable() || !schemaNamePattern.MatchString(groupID) || expectedGeneration <= 0 {
		return ErrAuthorityMismatch
	}
	var state, writer string
	var generation int64
	err := transaction.transaction.QueryRow(
		ctx, lockCutoverGroupSQL, tenantID, groupID,
	).Scan(&state, &writer, &generation)
	if err != nil || state != "v2" || writer != "v2" || generation != expectedGeneration {
		return ErrAuthorityMismatch
	}
	return nil
}
