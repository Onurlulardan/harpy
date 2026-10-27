package postgres

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type WorkPrincipalReader interface {
	Evidence(context.Context, workPrincipalTenantScope, string, uuid.UUID) (
		WorkPrincipalEvidence, error,
	)
	AvailableUsers(context.Context, workPrincipalTenantScope, string, int) (
		json.RawMessage, error,
	)
}

type workPrincipalReader struct{ port *workPrincipalPort }

func (reader *workPrincipalReader) Evidence(
	ctx context.Context, scope workPrincipalTenantScope,
	targetKind string, targetID uuid.UUID,
) (WorkPrincipalEvidence, error) {
	if reader == nil || reader.port == nil {
		return WorkPrincipalEvidence{}, ErrWorkPrincipalRead
	}
	return reader.port.evidence(ctx, scope, targetKind, targetID)
}

func (reader *workPrincipalReader) AvailableUsers(
	ctx context.Context, scope workPrincipalTenantScope, search string, limit int,
) (json.RawMessage, error) {
	if reader == nil || reader.port == nil {
		return nil, ErrWorkPrincipalRead
	}
	return reader.port.availableUsers(ctx, scope, search, limit)
}

var _ WorkPrincipalReader = (*workPrincipalReader)(nil)
