package customerlifecycle

import (
	"context"
	"errors"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

var (
	fixtureEnvironmentID = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	fixtureControlID     = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	fixtureTargetID      = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	fixtureActorID       = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	fixtureSessionID     = uuid.MustParse("55555555-5555-4555-8555-555555555555")
	fixtureNow           = time.Date(2026, 8, 2, 3, 23, 0, 0, time.UTC)
)

type fixtureClock struct{ now time.Time }

func (clock fixtureClock) Now() time.Time { return clock.now }

type fixturePolicy struct {
	mode tenantapi.DeploymentMode
	err  error
}

func (policy fixturePolicy) Snapshot() (tenantapi.DeploymentPolicy, error) {
	if policy.err != nil {
		return tenantapi.DeploymentPolicy{}, policy.err
	}
	mode := policy.mode
	if mode == "" {
		mode = tenantapi.DeploymentModeSaaS
	}
	var maximum *int64
	if mode == tenantapi.DeploymentModeOnPremise {
		zero := int64(0)
		maximum = &zero
	}
	return tenantapi.NewDeploymentPolicy(
		fixtureEnvironmentID.String(), mode, maximum, fixtureNow,
	)
}

func (policy fixturePolicy) PrimaryTarget() (tenantapi.PrimaryTenantTarget, error) {
	if policy.err != nil {
		return tenantapi.PrimaryTenantTarget{}, policy.err
	}
	mode, purpose := policy.mode, tenantapi.PrimaryTenantPurposeSaaS
	if mode == "" {
		mode = tenantapi.DeploymentModeSaaS
	}
	if mode == tenantapi.DeploymentModeOnPremise {
		purpose = tenantapi.PrimaryTenantPurposeOnPremise
	}
	return tenantapi.NewPrimaryTenantTarget(
		fixtureEnvironmentID.String(), mode, fixtureControlID.String(), purpose, "control",
	)
}

type fixtureAccess struct {
	requests []AccessRequest
	err      error
	now      time.Time
}

func (access *fixtureAccess) Verify(_ context.Context, request AccessRequest) (AccessDecision, error) {
	access.requests = append(access.requests, request)
	if access.err != nil {
		return AccessDecision{}, access.err
	}
	decision := AccessDecision{
		SubjectID: fixtureActorID, ControlTenantID: fixtureControlID,
		SessionID: fixtureSessionID, DeploymentMode: "saas", Capability: request.Capability,
	}
	if request.RequireStepUp {
		decision.StepUpAt = access.now.Add(-time.Minute)
	}
	return decision, nil
}

type fixtureStore struct {
	listRequest     ListStoreRequest
	mutationRequest MutationStoreRequest
	asyncRequest    AsyncStoreRequest
	page            Page
	tenant          Tenant
	mutation        MutationStoreResult
	async           AsyncStoreResult
	operation       Operation
	err             error
}

func (store *fixtureStore) List(_ context.Context, request ListStoreRequest) (Page, error) {
	store.listRequest = request
	return store.page, store.err
}

func (store *fixtureStore) Get(_ context.Context, _, _ uuid.UUID) (Tenant, error) {
	return store.tenant, store.err
}

func (store *fixtureStore) Mutate(_ context.Context, request MutationStoreRequest) (MutationStoreResult, error) {
	store.mutationRequest = request
	return store.mutation, store.err
}

func (store *fixtureStore) Accept(_ context.Context, request AsyncStoreRequest) (AsyncStoreResult, error) {
	store.asyncRequest = request
	return store.async, store.err
}

func (store *fixtureStore) ReadOperation(_ context.Context, _, _ uuid.UUID) (Operation, error) {
	return store.operation, store.err
}

type fixtureReadiness struct {
	proof PurgeProof
	err   error
	calls int
}

func (readiness *fixtureReadiness) Verify(_ context.Context, _ uuid.UUID) (PurgeProof, error) {
	readiness.calls++
	return readiness.proof, readiness.err
}

func validProof() PurgeProof {
	return PurgeProof{
		RetentionReceiptID:   uuid.MustParse("66666666-6666-4666-8666-666666666666"),
		BackupReceiptID:      uuid.MustParse("77777777-7777-4777-8777-777777777777"),
		ContributorReceiptID: uuid.MustParse("88888888-8888-4888-8888-888888888888"),
		EvidenceSHA256:       [32]byte{1}, VerifiedAt: fixtureNow.Add(-time.Hour),
	}
}

func newFixtureService(t *testing.T) (*Service, *fixtureAccess, *fixtureStore, *fixtureReadiness) {
	t.Helper()
	access := &fixtureAccess{now: fixtureNow}
	store := &fixtureStore{
		mutation: MutationStoreResult{Tenant: Tenant{ID: fixtureTargetID, Revision: 2}, Outcome: "applied"},
		async:    AsyncStoreResult{Operation: Operation{ID: fixtureSessionID, TargetTenantID: fixtureTargetID}, Outcome: "accepted"},
	}
	readiness := &fixtureReadiness{proof: validProof()}
	service, err := New(Dependencies{
		Policy: fixturePolicy{}, Access: access, Store: store,
		PurgeReadiness: readiness, Clock: fixtureClock{fixtureNow},
	})
	if err != nil {
		t.Fatal(err)
	}
	return service, access, store, readiness
}

func TestNewRejectsTypedNilDependencies(t *testing.T) {
	var access *fixtureAccess
	_, err := New(Dependencies{
		Policy: fixturePolicy{}, Access: access, Store: &fixtureStore{},
		PurgeReadiness: &fixtureReadiness{}, Clock: fixtureClock{fixtureNow},
	})
	if err == nil {
		t.Fatal("typed nil access accepted")
	}
	if !errors.Is(accessError(errors.New("unknown")), ErrUnavailable) {
		t.Fatal("unknown access error did not fail closed")
	}
}
