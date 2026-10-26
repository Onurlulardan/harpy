package initialadminsetup

import (
	"context"

	"github.com/google/uuid"
)

type FinalizeInput struct {
	SchemaName, ManifestSHA256 string
	OperationID, EnvironmentID uuid.UUID
	TenantID                   uuid.UUID
}

type Installation struct{ queryRow QueryRow }

func NewInstallation(queryRow QueryRow) (Installation, error) {
	if queryRow == nil {
		return Installation{}, ErrUnavailable
	}
	return Installation{queryRow: queryRow}, nil
}

func (installation Installation) Apply(
	ctx context.Context,
	input FinalizeInput,
) (uuid.UUID, error) {
	if installation.queryRow == nil || ctx == nil || ctx.Err() != nil {
		return uuid.Nil, ErrUnavailable
	}
	var artifactID uuid.UUID
	err := installation.queryRow(ctx, finalizeSQL, input.SchemaName,
		input.OperationID, input.EnvironmentID, input.TenantID,
		input.ManifestSHA256).Scan(&artifactID)
	return artifactID, err
}
