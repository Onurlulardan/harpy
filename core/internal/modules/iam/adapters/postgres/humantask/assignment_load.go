package humantask

import (
	"context"

	humantaskapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/humantask"
	"github.com/google/uuid"
)

func (reader *reader) LoadAssignment(
	ctx context.Context,
	scope tenantScope,
	request humantaskapi.AssignmentRequest,
) (humantaskapi.AssignmentEvidence, error) {
	if reader == nil || !reader.valid || request.Validate() != nil ||
		!scopeUsable(ctx, scope) || [16]byte(scope.TenantID()) != request.TenantID {
		return humantaskapi.AssignmentEvidence{}, errCandidateRead
	}
	bound, err := reader.bind(scope)
	if err != nil {
		return humantaskapi.AssignmentEvidence{}, errCandidateRead
	}
	var document []byte
	err = bound.database.QueryRow(ctx, resolveAssignmentSQL,
		scope.TenantID(), request.Assignment.Kind, uuid.UUID(request.Assignment.ID),
	).Scan(&document)
	if err != nil || len(document) == 0 {
		return humantaskapi.AssignmentEvidence{}, errCandidateRead
	}
	evidence, err := decodeAssignment(document, request)
	clear(document)
	if err != nil {
		return humantaskapi.AssignmentEvidence{}, errCandidateRead
	}
	return evidence, nil
}
