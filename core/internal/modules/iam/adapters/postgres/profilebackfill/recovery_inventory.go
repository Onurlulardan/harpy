package profilebackfill

import (
	"context"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

type RecoveryInventoryRow struct {
	Resource      profilebackfillapp.SourceResource
	SchemaVersion int64
	SchemaState   string
	LineageKind   string
}

func (row RecoveryInventoryRow) Validate() error {
	if row.Resource.Validate() != nil || row.SchemaVersion < 1 ||
		(row.LineageKind != "source" && row.LineageKind != "target") ||
		(row.LineageKind == "target" && row.SchemaState != "draft") ||
		(row.LineageKind == "source" && row.SchemaState != "active" &&
			row.SchemaState != "retired") {
		return profilebackfillapp.ErrInvalid
	}
	return nil
}

type RecoveryInventoryCursor interface {
	Fetch(context.Context) ([]RecoveryInventoryRow, error)
	Close(context.Context) error
}

func (reader *RecoveryReader) OpenInventory(
	ctx context.Context,
	scope tenantScope,
	parentID uuid.UUID,
	evidence profilebackfillapp.Evidence,
) (RecoveryInventoryCursor, error) {
	if !reader.usable() || parentID == uuid.Nil || evidence.Validate() != nil ||
		!scopeUsable(ctx, scope) || scope.ActorID() != evidence.ActorID {
		return nil, profilebackfillapp.ErrInvalid
	}
	bound, err := reader.value.bind(scope)
	if err != nil {
		return nil, err
	}
	var cursorName string
	err = bound.database.QueryRow(
		ctx, readRecoveryInventorySQL, parentID,
		evidence.ActorID, evidence.ManagePermissionID,
		evidence.ExpectedTokenVersion, evidence.ExpectedAuthorizationVersion,
	).Scan(&cursorName)
	if err != nil {
		return nil, recoveryStoreError(err)
	}
	if !validRecoveryInventoryCursorName(cursorName) {
		return nil, profilebackfillapp.ErrUnavailable
	}
	return &recoveryInventoryCursor{
		store: bound, scope: scope, name: cursorName,
	}, nil
}
