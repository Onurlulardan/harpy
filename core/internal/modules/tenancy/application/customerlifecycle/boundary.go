package customerlifecycle

import (
	"context"
	"errors"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type boundary struct {
	environmentID, controlTenantID uuid.UUID
}

func (service *Service) boundary() (boundary, error) {
	if service == nil {
		return boundary{}, ErrUnavailable
	}
	policy, err := service.dependencies.Policy.Snapshot()
	if err != nil || !policy.IsValid() {
		return boundary{}, ErrUnavailable
	}
	if policy.Mode() != tenantapi.DeploymentModeSaaS {
		return boundary{}, ErrModeDenied
	}
	target, err := service.dependencies.Policy.PrimaryTarget()
	if err != nil || !target.IsValid() || target.DeploymentMode() != policy.Mode() ||
		target.EnvironmentID() != policy.EnvironmentID() ||
		target.Purpose() != tenantapi.PrimaryTenantPurposeSaaS {
		return boundary{}, ErrUnavailable
	}
	environmentID, environmentErr := uuid.Parse(policy.EnvironmentID())
	controlID, controlErr := uuid.Parse(target.TenantID())
	if environmentErr != nil || controlErr != nil || environmentID == uuid.Nil || controlID == uuid.Nil {
		return boundary{}, ErrUnavailable
	}
	return boundary{environmentID: environmentID, controlTenantID: controlID}, nil
}

func (service *Service) authorize(ctx context.Context, edge boundary, credential CredentialEnvelope,
	capability string, target uuid.UUID, requireStepUp bool, purpose [32]byte,
) (AccessDecision, error) {
	decision, err := service.dependencies.Access.Verify(ctx, AccessRequest{
		Credential: credential, Capability: capability, ExpectedMode: "saas",
		TargetTenantID: target, RequireStepUp: requireStepUp, StepUpPurpose: purpose,
	})
	if err != nil {
		return AccessDecision{}, err
	}
	if decision.SubjectID == uuid.Nil || decision.ControlTenantID != edge.controlTenantID ||
		decision.SessionID == uuid.Nil || decision.DeploymentMode != "saas" ||
		decision.Capability != capability {
		return AccessDecision{}, ErrForbidden
	}
	if requireStepUp {
		now := service.dependencies.Clock.Now().UTC().Truncate(time.Second)
		if decision.StepUpAt.IsZero() || decision.StepUpAt.Location() != time.UTC ||
			decision.StepUpAt.After(now) || now.Sub(decision.StepUpAt) > 10*time.Minute {
			return AccessDecision{}, ErrForbidden
		}
	} else if !decision.StepUpAt.IsZero() {
		return AccessDecision{}, ErrForbidden
	}
	return decision, nil
}

func accessError(err error) error {
	switch {
	case errors.Is(err, ErrAuthentication), errors.Is(err, ErrForbidden),
		errors.Is(err, ErrUnavailable):
		return err
	default:
		return ErrUnavailable
	}
}
