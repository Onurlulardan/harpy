package iam

import (
	"context"

	humantaskapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/humantask"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HumanTaskAccessRuntime) VerifyHumanTaskEvidence(
	ctx context.Context,
	scope platformtx.TenantScope,
	request humantaskapi.VerificationRequest,
) (humantaskapi.VerificationEvidence, error) {
	if !runtime.Valid() || request.Validate() != nil {
		return humantaskapi.VerificationEvidence{}, humantaskapi.ErrInvalidVerification
	}
	evidence, err := runtime.verifier.Verify(ctx, scope, request)
	if err != nil || !evidence.Matches(request) {
		if err == nil {
			err = humantaskapi.ErrInvalidVerification
		}
		return humantaskapi.VerificationEvidence{}, err
	}
	return evidence, nil
}

func (runtime *HumanTaskAccessRuntime) VerifyHumanTaskAssignment(
	ctx context.Context,
	scope platformtx.TenantScope,
	request humantaskapi.AssignmentRequest,
) (humantaskapi.AssignmentEvidence, error) {
	if !runtime.Valid() || request.Validate() != nil {
		return humantaskapi.AssignmentEvidence{}, humantaskapi.ErrInvalidVerification
	}
	evidence, err := runtime.verifier.VerifyAssignment(ctx, scope, request)
	if err != nil || !evidence.Matches(request) {
		if err == nil {
			err = humantaskapi.ErrInvalidVerification
		}
		return humantaskapi.AssignmentEvidence{}, err
	}
	return evidence, nil
}
