package api

import (
	"context"
	"errors"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformopsmodule "github.com/Onurlulardan/harpy/core/internal/modules/platformops"
)

type platformOpsIAMAccess interface {
	Verify(context.Context, iammodule.PlatformAccessRequest) (
		iammodule.PlatformAccessDecision, error)
}

type platformOpsAccessVerifier struct{ iam platformOpsIAMAccess }

func newPlatformOpsAccessVerifier(
	iam platformOpsIAMAccess,
) (*platformOpsAccessVerifier, error) {
	if iam == nil {
		return nil, fmt.Errorf("PlatformOps IAM access runtime is required")
	}
	return &platformOpsAccessVerifier{iam: iam}, nil
}

func (verifier *platformOpsAccessVerifier) Verify(
	ctx context.Context,
	request platformopsmodule.HTTPAccessRequest,
) (platformopsmodule.HTTPAccess, error) {
	if verifier == nil || verifier.iam == nil {
		return platformopsmodule.HTTPAccess{}, platformopsmodule.ErrHTTPUnavailable
	}
	decision, err := verifier.iam.Verify(ctx, iammodule.PlatformAccessRequest{
		AuthorizationValues: request.AuthorizationValues,
		TenantIDValues:      request.TenantIDValues, TenantSlugValues: request.TenantSlugValues,
		Capability: request.Capability, ExpectedMode: request.ExpectedMode,
		RequireStepUp: request.RequireStepUp, StepUpPurpose: request.StepUpPurpose,
	})
	if err != nil {
		return platformopsmodule.HTTPAccess{}, mapPlatformOpsAccessError(err)
	}
	return platformopsmodule.HTTPAccess{ActorID: decision.SubjectID.String(),
		TenantID: decision.TenantID.String(), SessionID: decision.SessionID.String(),
		TokenVersion:         decision.TokenVersion,
		AuthorizationVersion: decision.AuthorizationVersion,
		PlatformClock:        decision.PlatformClock}, nil
}

func mapPlatformOpsAccessError(err error) error {
	switch {
	case errors.Is(err, iammodule.ErrPlatformAccessAuthentication):
		return platformopsmodule.ErrHTTPAuthentication
	case errors.Is(err, iammodule.ErrPlatformAccessForbidden):
		return platformopsmodule.ErrHTTPForbidden
	default:
		return platformopsmodule.ErrHTTPUnavailable
	}
}

var _ platformopsmodule.HTTPAccessVerifier = (*platformOpsAccessVerifier)(nil)
