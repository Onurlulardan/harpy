package initialadminsetup

import (
	"context"

	"github.com/google/uuid"
)

type ReissueInput struct {
	OperationID, ReissueOperationID, TenantID uuid.UUID
}

func (installation Installation) Reissue(
	ctx context.Context,
	input ReissueInput,
) (uuid.UUID, error) {
	if installation.queryRow == nil || ctx == nil || ctx.Err() != nil ||
		input.OperationID == uuid.Nil || input.ReissueOperationID == uuid.Nil ||
		input.TenantID == uuid.Nil {
		return uuid.Nil, ErrUnavailable
	}
	var artifactID uuid.UUID
	err := installation.queryRow(ctx, reissueSQL, input.OperationID,
		input.ReissueOperationID, input.TenantID).Scan(&artifactID)
	return artifactID, err
}
