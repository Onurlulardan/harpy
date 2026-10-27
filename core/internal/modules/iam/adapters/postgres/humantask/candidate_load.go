package humantask

import (
	"context"

	humantaskapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/humantask"
	humantaskapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/humantask"
	"github.com/google/uuid"
)

func (reader *reader) LoadCandidate(
	ctx context.Context,
	scope tenantScope,
	request humantaskapi.VerificationRequest,
) (humantaskapp.CandidateEvidence, error) {
	if reader == nil || !reader.valid || request.Validate() != nil ||
		!scopeUsable(ctx, scope) || [16]byte(scope.TenantID()) != request.TenantID {
		return humantaskapp.CandidateEvidence{}, errCandidateRead
	}
	bound, err := reader.bind(scope)
	if err != nil {
		return humantaskapp.CandidateEvidence{}, errCandidateRead
	}
	var document []byte
	err = bound.database.QueryRow(ctx, resolveCandidateSQL,
		scope.TenantID(), uuid.UUID(request.ActorID), request.Candidate.Kind,
		uuid.UUID(request.Candidate.ID),
	).Scan(&document)
	if err != nil || len(document) == 0 {
		return humantaskapp.CandidateEvidence{}, errCandidateRead
	}
	evidence, err := decodeCandidate(document, request)
	clear(document)
	if err != nil {
		return humantaskapp.CandidateEvidence{}, errCandidateRead
	}
	return evidence, nil
}
