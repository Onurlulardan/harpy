package profilebackfill

import (
	"context"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func (reader *RecoveryReader) InspectParent(
	ctx context.Context,
	scope tenantScope,
	parentID uuid.UUID,
	evidence profilebackfillapp.Evidence,
) (RecoveryContract, error) {
	if !reader.usable() || parentID == uuid.Nil || evidence.Validate() != nil ||
		!scopeUsable(ctx, scope) || scope.ActorID() != evidence.ActorID {
		return RecoveryContract{}, profilebackfillapp.ErrInvalid
	}
	bound, err := reader.value.bind(scope)
	if err != nil {
		return RecoveryContract{}, err
	}
	var row recoveryParentRow
	err = bound.database.QueryRow(
		ctx, readRecoveryParentSQL, parentID, evidence.ActorID,
		evidence.ManagePermissionID, evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion,
	).Scan(
		&row.ID, &row.RecoveryRootPlanID, &row.RecoverySequence,
		&row.PlanSHA256, &row.ResourceKind, &row.SourceSchemaID,
		&row.TargetSchemaID, &row.ConverterVersion, &row.ExistingChildPlanID,
		&row.SourceSnapshot, &row.TargetSnapshot, &row.Mappings,
	)
	if err != nil {
		return RecoveryContract{}, profilebackfillapp.ErrUnavailable
	}
	if row.ExistingChildPlanID != nil && row.ID == parentID &&
		*row.ExistingChildPlanID != uuid.Nil {
		return RecoveryContract{ExistingChildPlanID: *row.ExistingChildPlanID}, nil
	}
	contract, err := decodeRecoveryContract(row)
	if err != nil || contract.Parent.ID != parentID {
		return RecoveryContract{}, profilebackfillapp.ErrUnavailable
	}
	return contract, nil
}
