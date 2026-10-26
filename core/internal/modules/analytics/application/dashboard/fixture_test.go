package dashboard

import (
	"context"
	"time"

	analyticsdashboard "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dashboard"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var (
	dashboardFixtureOperation = uuid.MustParse("018f0000-0000-7000-8000-000000000201")
	dashboardFixtureID        = uuid.MustParse("018f0000-0000-7000-8000-000000000202")
	dashboardFixtureVersion   = uuid.MustParse("018f0000-0000-7000-8000-000000000203")
	dashboardFixtureNext      = uuid.MustParse("018f0000-0000-7000-8000-000000000204")
	dashboardFixtureActor     = uuid.MustParse("018f0000-0000-7000-8000-000000000205")
	dashboardFixtureTenant    = uuid.MustParse("018f0000-0000-7000-8000-000000000206")
	dashboardFixturePrincipal = uuid.MustParse("018f0000-0000-7000-8000-000000000207")
	dashboardFixtureNow       = time.Date(2026, 7, 18, 11, 0, 0, 123000, time.UTC)
)

type dashboardFixtureStore struct {
	create     func(context.Context, CreateMutation) (Snapshot, bool, error)
	get        func(context.Context, uuid.UUID) (Snapshot, error)
	list       func(context.Context, int, *uuid.UUID) (Page, error)
	update     func(context.Context, UpdateMutation) (Snapshot, bool, error)
	delete     func(context.Context, DeleteMutation) (Snapshot, bool, error)
	publish    func(context.Context, PublishMutation) (Snapshot, bool, error)
	grant      func(context.Context, ACLMutation) (ACL, bool, error)
	revoke     func(context.Context, ACLMutation) (bool, error)
	setDefault func(context.Context, DefaultMutation) (bool, error)
}

func (s *dashboardFixtureStore) Create(c context.Context, m CreateMutation) (Snapshot, bool, error) {
	return s.create(c, m)
}
func (s *dashboardFixtureStore) Get(c context.Context, id uuid.UUID) (Snapshot, error) {
	return s.get(c, id)
}

func (s *dashboardFixtureStore) ReadDefault(c context.Context, surface string) (Snapshot, error) {
	return s.Get(c, dashboardFixtureID)
}
func (s *dashboardFixtureStore) List(c context.Context, n int, id *uuid.UUID) (Page, error) {
	return s.list(c, n, id)
}
func (s *dashboardFixtureStore) Update(c context.Context, m UpdateMutation) (Snapshot, bool, error) {
	return s.update(c, m)
}
func (s *dashboardFixtureStore) Delete(c context.Context, m DeleteMutation) (Snapshot, bool, error) {
	return s.delete(c, m)
}
func (s *dashboardFixtureStore) Publish(c context.Context, m PublishMutation) (Snapshot, bool, error) {
	return s.publish(c, m)
}
func (s *dashboardFixtureStore) GrantACL(c context.Context, m ACLMutation) (ACL, bool, error) {
	return s.grant(c, m)
}
func (s *dashboardFixtureStore) RevokeACL(c context.Context, m ACLMutation) (bool, error) {
	return s.revoke(c, m)
}
func (s *dashboardFixtureStore) SetDefault(c context.Context, m DefaultMutation) (bool, error) {
	return s.setDefault(c, m)
}

type dashboardFixtureIDs struct {
	values []uuid.UUID
	index  int
}

func (source *dashboardFixtureIDs) NewUUIDv7() (uuid.UUID, error) {
	if source.index >= len(source.values) {
		return uuid.Nil, ErrUnavailable
	}
	value := source.values[source.index]
	source.index++
	return value, nil
}

type dashboardFixtureClock struct{}

func (dashboardFixtureClock) Now() time.Time { return dashboardFixtureNow }

type dashboardFixtureEntitlements struct {
	decision billingapi.MeteredResourceMutationDecision
	err      error
	called   bool
	request  billingapi.MeteredResourceMutationRequest
}

func (fixture *dashboardFixtureEntitlements) DecideMeteredResourceMutation(
	_ context.Context, _ platformtx.TenantScope,
	request billingapi.MeteredResourceMutationRequest,
) (billingapi.MeteredResourceMutationDecision, error) {
	fixture.called, fixture.request = true, request
	if fixture.err != nil {
		return billingapi.MeteredResourceMutationDecision{}, fixture.err
	}
	decision := fixture.decision
	if !decision.Resource.Valid() {
		decision = dashboardFixtureMutationDecision(request.Mutation, false)
		decision.Resource = request.Resource
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

func dashboardFixtureMutationDecision(
	mutation billingapi.MeteredResourceMutation, replay bool,
) billingapi.MeteredResourceMutationDecision {
	decision := billingapi.MeteredResourceMutationDecision{
		Resource: billingapi.MeteredResourceAnalyticsDashboard, Mutation: mutation,
		TenantID: dashboardFixtureTenant, OperationID: dashboardFixtureOperation,
		OwnerFactID: dashboardFixtureID, Allowed: true, Replay: replay,
	}
	switch mutation {
	case billingapi.MeteredResourceCreate, billingapi.MeteredResourceRestore:
		decision.FeatureEnabled, decision.GaugeDelta, decision.Unlimited = true, 1, true
	case billingapi.MeteredResourceUpdate, billingapi.MeteredResourcePublish,
		billingapi.MeteredResourceRevalidate:
		decision.FeatureEnabled = true
	case billingapi.MeteredResourceDelete:
		decision.GaugeDelta, decision.Current = -1, 1
	}
	return decision
}

type dashboardFixtureQueries struct {
	manifest [32]byte
	err      error
}

func (q dashboardFixtureQueries) VerifyPublished(context.Context, platformtx.TenantScope, []uuid.UUID) ([32]byte, error) {
	return q.manifest, q.err
}

func dashboardFixtureSnapshot(published bool) Snapshot {
	layout, _ := NewDocument([]byte(`{}`))
	manifest := [32]byte{1}
	return Snapshot{DashboardID: dashboardFixtureID, VersionID: dashboardFixtureVersion,
		VersionNumber: 1, RootRevision: 1, VersionRevision: 1,
		Lifecycle: analyticsdashboard.LifecycleActive, Visibility: analyticsdashboard.VisibilityPrivate,
		Published: published, Name: "Operations", LayoutDocument: layout.CanonicalJSON(),
		LayoutSHA256: layout.SHA256(), ManifestSHA256: manifest,
		Widgets: []Widget{{WidgetID: dashboardFixtureNext, QueryVersionID: dashboardFixtureVersion,
			Position: 0, ParameterDocument: []byte(`{}`), OutputDocument: []byte(`{}`)}},
		CreatedAt: dashboardFixtureNow, UpdatedAt: dashboardFixtureNow}
}
func dashboardFixtureIdentity() ScopeIdentity {
	return ScopeIdentity{dashboardFixtureActor, dashboardFixtureTenant}
}
