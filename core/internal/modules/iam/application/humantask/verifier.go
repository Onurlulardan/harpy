package humantask

import (
	"context"
	"reflect"

	humantaskapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/humantask"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const humanTaskResource = "workflow.human_tasks"

type Verifier struct {
	authorizer  CurrentPermissionPort
	candidates  CandidatePort
	assignments AssignmentPort
	valid       bool
}

func NewVerifier(
	authorizer CurrentPermissionPort,
	candidates CandidatePort,
	assignments AssignmentPort,
) (Verifier, error) {
	if currentPermissionPortNil(authorizer) || candidatePortNil(candidates) ||
		assignmentPortNil(assignments) {
		return Verifier{}, ErrInvalid
	}
	return Verifier{authorizer: authorizer, candidates: candidates,
		assignments: assignments, valid: true}, nil
}

func (verifier Verifier) VerifyAssignment(
	ctx context.Context,
	scope platformtx.TenantScope,
	request humantaskapi.AssignmentRequest,
) (humantaskapi.AssignmentEvidence, error) {
	if !verifier.valid || assignmentPortNil(verifier.assignments) ||
		request.Validate() != nil || !assignmentScopeUsable(ctx, scope, request) {
		return humantaskapi.AssignmentEvidence{}, ErrInvalid
	}
	evidence, err := verifier.assignments.LoadAssignment(ctx, scope, request)
	if err != nil || !evidence.Matches(request) {
		return humantaskapi.AssignmentEvidence{}, ErrForbidden
	}
	return evidence, nil
}

func (verifier Verifier) Verify(
	ctx context.Context,
	scope platformtx.TenantScope,
	request humantaskapi.VerificationRequest,
) (humantaskapi.VerificationEvidence, error) {
	if !verifier.valid || candidatePortNil(verifier.candidates) ||
		request.Validate() != nil || !scopeUsable(ctx, scope, request) {
		return humantaskapi.VerificationEvidence{}, ErrInvalid
	}
	candidate, err := verifier.candidates.LoadCandidate(ctx, scope, request)
	if err != nil || !candidate.valid(request) {
		return humantaskapi.VerificationEvidence{}, ErrForbidden
	}
	permission, err := verifier.authorizer.ResolveCurrentHumanPermission(
		ctx, scope, request.ActorID, request.Action,
	)
	if err != nil || !permission.valid() ||
		permission.AuthorizationVersion != candidate.AuthorizationVersion {
		return humantaskapi.VerificationEvidence{}, ErrForbidden
	}
	evidence := humantaskapi.VerificationEvidence{
		TenantID: request.TenantID, ActorID: request.ActorID,
		AuthorizationVersion: candidate.AuthorizationVersion,
		Candidate:            request.Candidate, Action: request.Action,
		PermissionID: permission.PermissionID, VerifiedAt: request.VerifyAt,
	}
	evidence.EvidenceSHA256 = evidenceDigest(evidence, candidate.EvidenceSHA256)
	if !evidence.Matches(request) {
		return humantaskapi.VerificationEvidence{}, ErrUnavailable
	}
	return evidence, nil
}

func currentPermissionPortNil(port CurrentPermissionPort) bool {
	if port == nil {
		return true
	}
	value := reflect.ValueOf(port)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func candidatePortNil(port CandidatePort) bool {
	if port == nil {
		return true
	}
	value := reflect.ValueOf(port)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func assignmentPortNil(port AssignmentPort) bool {
	if port == nil {
		return true
	}
	value := reflect.ValueOf(port)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
