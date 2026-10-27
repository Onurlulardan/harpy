package environmentoperator

import (
	"context"
	"strings"
	"time"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type policyFixture struct{ createdAt time.Time }

func (fixture policyFixture) CreateDeploymentPolicy(_ context.Context,
	request tenantapi.InitializeDeploymentPolicyRequest,
) (tenantapi.DeploymentPolicyInitialization, error) {
	var maximum *int64
	if request.DeploymentMode == tenantapi.DeploymentModeOnPremise {
		value := int64(0)
		maximum = &value
	}
	policy, err := tenantapi.NewDeploymentPolicy(
		tenantapi.UUIDString(request.EnvironmentID), request.DeploymentMode,
		maximum, fixture.createdAt)
	return tenantapi.DeploymentPolicyInitialization{OperationID: request.OperationID,
		Policy: policy, DatabaseNow: fixture.createdAt, Replayed: false}, err
}

type offeringFixture struct{}

func (offeringFixture) Resolve(_ context.Context,
	request OfferingRequest,
) (tenantapi.BaseOfferingPin, error) {
	pin := tenantapi.BaseOfferingPin{VersionID: stableID(request.Manifest, "offering"),
		Revision: 1, SHA256: strings.Repeat("4", 64),
		EffectiveAt: request.CreatedAt}
	if request.Manifest.DeploymentMode() == installationv1.DeploymentOnPremise {
		pin.Source = tenantapi.OfferingSourcePin{Present: true, Sequence: 1,
			SHA256: strings.Repeat("5", 64)}
	}
	return pin, nil
}

type authorityFixture struct{ prepares, acknowledgements int }

func (fixture *authorityFixture) Prepare(_ context.Context,
	request AuthorityRequest,
) (tenantapi.ProtectedAuthorityPlan, error) {
	fixture.prepares++
	return tenantapi.ProtectedAuthorityPlan{RegistryBindingID: request.RegistryBindingID,
		ScopeKind: tenantapi.AuthorityScopeEnvironment, ScopeID: request.EnvironmentID,
		Generation: 1, StateSHA256: strings.Repeat("6", 64),
		ProposalSHA256: strings.Repeat("7", 64)}, nil
}

func (fixture *authorityFixture) Acknowledge(_ context.Context,
	request AuthorityRequest,
) (tenantapi.AuthorityAcknowledgement, error) {
	fixture.acknowledgements++
	return tenantapi.AuthorityAcknowledgement{OperationID: request.OperationID,
		RegistryBindingID: request.RegistryBindingID,
		ScopeKind:         tenantapi.AuthorityScopeEnvironment, ScopeID: request.EnvironmentID,
		Generation: 1, StateSHA256: strings.Repeat("6", 64),
		ProposalSHA256: strings.Repeat("7", 64), AckSHA256: strings.Repeat("8", 64),
		RegistrySequence: 1, VerifiedAt: request.IssuedAt}, nil
}

type finalizerFixture struct {
	onboarding *onboardingFixture
	calls      int
	err        error
}

func (fixture *finalizerFixture) Finalize(_ context.Context,
	request FinalizationRequest,
) error {
	fixture.calls++
	if fixture.err != nil {
		return fixture.err
	}
	if request.ExpectedRevision != fixture.onboarding.revision ||
		request.AttemptID == uuid.Nil || request.LeaseTokenSHA256 == "" {
		return ErrFenced
	}
	fixture.onboarding.revision++
	fixture.onboarding.state = tenantapi.OnboardingSucceeded
	return nil
}

func recoveryFixture() RecoverySigningResult {
	return RecoverySigningResult{Generation: 1,
		TrustBundleSHA256: strings.Repeat("3", 64)}
}
