package humantask

import (
	"context"

	humantaskapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/humantask"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type CandidateEvidence struct {
	ActorID              [16]byte
	AuthorizationVersion int64
	Candidate            humantaskapi.Candidate
	EvidenceSHA256       [32]byte
}

func (evidence CandidateEvidence) valid(request humantaskapi.VerificationRequest) bool {
	return evidence.ActorID == request.ActorID && evidence.AuthorizationVersion > 0 &&
		evidence.Candidate == request.Candidate && evidence.EvidenceSHA256 != [32]byte{}
}

type CandidatePort interface {
	LoadCandidate(
		context.Context,
		platformtx.TenantScope,
		humantaskapi.VerificationRequest,
	) (CandidateEvidence, error)
}

type AssignmentPort interface {
	LoadAssignment(
		context.Context,
		platformtx.TenantScope,
		humantaskapi.AssignmentRequest,
	) (humantaskapi.AssignmentEvidence, error)
}

type EvidencePort interface {
	CandidatePort
	AssignmentPort
}
