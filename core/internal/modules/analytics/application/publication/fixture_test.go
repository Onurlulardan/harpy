package publication

import (
	"context"
	"crypto/sha256"
	"time"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var (
	publicationFixtureRelease   = uuid.MustParse("018f0000-0000-7000-8000-000000000301")
	publicationFixtureVersion   = uuid.MustParse("018f0000-0000-7000-8000-000000000302")
	publicationFixtureTenant    = uuid.MustParse("018f0000-0000-7000-8000-000000000303")
	publicationFixtureActor     = uuid.MustParse("018f0000-0000-7000-8000-000000000304")
	publicationFixtureApprover  = uuid.MustParse("018f0000-0000-7000-8000-000000000305")
	publicationFixtureOperation = uuid.MustParse("018f0000-0000-7000-8000-000000000306")
	publicationFixtureNow       = time.Date(2026, 7, 18, 12, 0, 0, 123000, time.UTC)
	publicationFixturePublicID  = [16]byte{1, 2, 3, 4}
	publicationFixtureToken     = []byte("0123456789abcdef0123456789abcdef")
)

type publicationFixtureClock struct{ value time.Time }

func (clock publicationFixtureClock) Now() time.Time { return clock.value }

type publicationFixtureEntitlements struct {
	decision billingapi.MeteredResourceMutationDecision
	err      error
	called   bool
	request  billingapi.MeteredResourceMutationRequest
}

func (fixture *publicationFixtureEntitlements) DecideMeteredResourceMutation(
	_ context.Context, _ platformtx.TenantScope,
	request billingapi.MeteredResourceMutationRequest,
) (billingapi.MeteredResourceMutationDecision, error) {
	fixture.called, fixture.request = true, request
	if fixture.err != nil {
		return billingapi.MeteredResourceMutationDecision{}, fixture.err
	}
	decision := fixture.decision
	if !decision.Resource.Valid() {
		decision = billingapi.MeteredResourceMutationDecision{
			Resource: request.Resource, Mutation: request.Mutation,
			FeatureEnabled: true, Allowed: true,
		}
	}
	decision.TenantID, decision.ActorID = request.TenantID, request.ActorID
	decision.IdempotencyKey, decision.RequestSHA256 = request.IdempotencyKey, request.RequestSHA256
	if decision.OperationID == uuid.Nil {
		decision.OperationID = request.ProposedOperationID
	}
	if decision.OwnerFactID == uuid.Nil {
		decision.OwnerFactID = request.ProposedOwnerFactID
	}
	return decision, nil
}

type publicationFixtureAuthorizer struct {
	stepUp    time.Time
	stepUpErr error
}

func (*publicationFixtureAuthorizer) Authorize(context.Context, platformtx.TenantScope, Permission) error {
	return nil
}
func (authorizer *publicationFixtureAuthorizer) FreshStepUp(context.Context, platformtx.TenantScope, time.Time) (time.Time, error) {
	return authorizer.stepUp, authorizer.stepUpErr
}

type publicationFixtureDirectory struct {
	route PublicRoute
	err   error
}

func (directory publicationFixtureDirectory) Route(context.Context, [16]byte) (PublicRoute, error) {
	return directory.route, directory.err
}

type publicationFixtureScopes struct{ called bool }

func (scopes *publicationFixtureScopes) InPublicationSnapshot(ctx context.Context, _ PublicRoute,
	execute func(context.Context, platformtx.TenantScope) error,
) error {
	scopes.called = true
	return execute(ctx, platformtx.TenantScope{})
}

type publicationFixtureExecutor struct {
	called bool
	result PublicResult
}

func (executor *publicationFixtureExecutor) Execute(context.Context, platformtx.TenantScope, PublicRoute, []byte) (PublicResult, error) {
	executor.called = true
	return executor.result, nil
}

type publicationFixtureLease struct{ released bool }

func (lease *publicationFixtureLease) Close(context.Context) error {
	lease.released = true
	return nil
}

type publicationFixtureLimiter struct {
	called    bool
	admission PublicAdmission
	lease     *publicationFixtureLease
	err       error
}

func (limiter *publicationFixtureLimiter) Admit(_ context.Context, admission PublicAdmission) (PublicLease, error) {
	limiter.called, limiter.admission = true, admission
	return limiter.lease, limiter.err
}

func publicationFixtureRoute(audience analyticspublication.Audience) PublicRoute {
	digest := sha256.Sum256(publicationFixtureToken)
	expires := publicationFixtureNow.Add(time.Hour)
	return PublicRoute{TenantID: publicationFixtureTenant, ReleaseID: publicationFixtureRelease,
		ResourceKind: ResourceQuery, ResourceVersionID: publicationFixtureVersion,
		Audience: audience, State: analyticspublication.StateActive, Generation: 4,
		PublicID: publicationFixturePublicID, CapabilitySHA256: digest,
		CapabilityExpiresAt: &expires}
}
