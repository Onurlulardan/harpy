package query

import (
	"context"
	"time"

	analyticsforms "github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/formsreference"
	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var (
	queryFixtureOperation = uuid.MustParse("018f0000-0000-7000-8000-000000000101")
	queryFixtureID        = uuid.MustParse("018f0000-0000-7000-8000-000000000102")
	queryFixtureVersion   = uuid.MustParse("018f0000-0000-7000-8000-000000000103")
	queryFixtureNext      = uuid.MustParse("018f0000-0000-7000-8000-000000000104")
	queryFixtureActor     = uuid.MustParse("018f0000-0000-7000-8000-000000000105")
	queryFixtureTenant    = uuid.MustParse("018f0000-0000-7000-8000-000000000106")
	queryFixtureNow       = time.Date(2026, 7, 18, 10, 0, 0, 123000, time.UTC)
)

type queryFixtureStore struct {
	create          func(context.Context, CreateMutation) (Snapshot, bool, error)
	get             func(context.Context, uuid.UUID) (Snapshot, error)
	latestPublished func(context.Context, uuid.UUID) (Snapshot, error)
	resolveQueryID  func(context.Context, uuid.UUID) (uuid.UUID, error)
	getVersion      func(context.Context, uuid.UUID, uuid.UUID) (Snapshot, error)
	dependencies    func(context.Context, uuid.UUID, uuid.UUID) ([]DependencySnapshot, error)
	list            func(context.Context, int, *uuid.UUID) (Page, error)
	update          func(context.Context, UpdateMutation) (Snapshot, bool, error)
	delete          func(context.Context, DeleteMutation) (Snapshot, bool, error)
	publish         func(context.Context, PublishMutation) (Snapshot, bool, error)
}

func (s *queryFixtureStore) Create(c context.Context, m CreateMutation) (Snapshot, bool, error) {
	return s.create(c, m)
}
func (s *queryFixtureStore) Get(c context.Context, id uuid.UUID) (Snapshot, error) {
	return s.get(c, id)
}
func (s *queryFixtureStore) LatestPublished(c context.Context, id uuid.UUID) (Snapshot, error) {
	return s.latestPublished(c, id)
}
func (s *queryFixtureStore) ResolveQueryID(c context.Context, version uuid.UUID) (uuid.UUID, error) {
	if s.resolveQueryID == nil {
		return queryFixtureID, nil
	}
	return s.resolveQueryID(c, version)
}
func (s *queryFixtureStore) LoadVersion(c context.Context, id, version uuid.UUID) (Snapshot, error) {
	return s.getVersion(c, id, version)
}
func (s *queryFixtureStore) Dependencies(c context.Context, id, version uuid.UUID) ([]DependencySnapshot, error) {
	return s.dependencies(c, id, version)
}
func (s *queryFixtureStore) List(c context.Context, limit int, after *uuid.UUID) (Page, error) {
	return s.list(c, limit, after)
}
func (s *queryFixtureStore) Update(c context.Context, m UpdateMutation) (Snapshot, bool, error) {
	return s.update(c, m)
}
func (s *queryFixtureStore) Delete(c context.Context, m DeleteMutation) (Snapshot, bool, error) {
	return s.delete(c, m)
}
func (s *queryFixtureStore) Publish(c context.Context, m PublishMutation) (Snapshot, bool, error) {
	return s.publish(c, m)
}

type queryFixtureIDs struct {
	values []uuid.UUID
	index  int
}

func (source *queryFixtureIDs) NewUUIDv7() (uuid.UUID, error) {
	if source.index >= len(source.values) {
		return uuid.Nil, ErrUnavailable
	}
	value := source.values[source.index]
	source.index++
	return value, nil
}

type queryFixtureClock struct{ value time.Time }

func (clock queryFixtureClock) Now() time.Time { return clock.value }

func queryFixtureApplication(ids ...uuid.UUID) *Application {
	return &Application{ids: &queryFixtureIDs{values: ids}, clock: queryFixtureClock{queryFixtureNow},
		references: queryFixtureReferences{}, entitlements: &queryFixtureEntitlements{}}
}

func queryFixtureMutationDecision(
	mutation billingapi.MeteredResourceMutation, replay bool,
) billingapi.MeteredResourceMutationDecision {
	decision := billingapi.MeteredResourceMutationDecision{
		Resource: billingapi.MeteredResourceAPIEndpoint, Mutation: mutation,
		TenantID: queryFixtureTenant, OperationID: queryFixtureOperation,
		OwnerFactID: queryFixtureID, Allowed: true, Replay: replay,
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

type queryFixtureEntitlements struct {
	decision billingapi.MeteredResourceMutationDecision
	err      error
	called   bool
	request  billingapi.MeteredResourceMutationRequest
}

func (fixture *queryFixtureEntitlements) DecideMeteredResourceMutation(
	_ context.Context, _ platformtx.TenantScope,
	request billingapi.MeteredResourceMutationRequest,
) (billingapi.MeteredResourceMutationDecision, error) {
	fixture.called, fixture.request = true, request
	if fixture.err != nil {
		return billingapi.MeteredResourceMutationDecision{}, fixture.err
	}
	decision := fixture.decision
	if !decision.Resource.Valid() {
		decision = queryFixtureMutationDecision(request.Mutation, false)
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

type queryFixtureReferences struct{}

func (queryFixtureReferences) BindVersion(
	context.Context, platformtx.TenantScope, analyticsforms.Source,
	[]analyticsforms.Target, time.Time,
) error {
	return nil
}

func (queryFixtureReferences) DecommissionSource(
	context.Context, platformtx.TenantScope, analyticsforms.ID,
	*analyticsforms.ID, time.Time,
) error {
	return nil
}
func queryFixtureIdentity() ScopeIdentity {
	return ScopeIdentity{queryFixtureActor, queryFixtureTenant}
}
func queryFixtureDocument(value string) Document {
	document, _ := NewDocument([]byte(value))
	return document
}
func queryFixtureSnapshot(state analyticsquery.VersionState, policy analyticspublication.CachePolicy) Snapshot {
	manifest := [32]byte{1}
	return Snapshot{QueryID: queryFixtureID, VersionID: queryFixtureVersion,
		VersionNumber: 1, RootRevision: 1, ReferenceEpoch: 1, VersionRevision: 1,
		Lifecycle: analyticsquery.LifecycleActive, VersionState: state, CachePolicy: policy,
		Name: "Orders", QueryDocument: []byte(`{}`), ParameterDocument: []byte(`{}`),
		OutputDocument: []byte(`{}`), ManifestSHA256: manifest,
		CreatedAt: queryFixtureNow, UpdatedAt: queryFixtureNow}
}

type queryFixtureContracts struct {
	binding ExecutionBinding
	result  ExecutionResult
}

func (p queryFixtureContracts) Normalize(context.Context, platformtx.TenantScope, Document, Document, Document, []DependencySnapshot) (NormalizedContract, error) {
	d := queryFixtureDocument(`{}`)
	return NormalizedContract{d, d, d}, nil
}
func (p queryFixtureContracts) BindExecution(context.Context, platformtx.TenantScope, Snapshot, []DependencySnapshot, Document) (ExecutionBinding, error) {
	return p.binding, nil
}
func (p queryFixtureContracts) Run(context.Context, platformtx.TenantScope, Snapshot, []DependencySnapshot, Document) (ExecutionResult, error) {
	return p.result, nil
}
