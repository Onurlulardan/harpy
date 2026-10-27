package onboarding

import (
	"context"
	"strings"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy/domain"
	"github.com/google/uuid"
)

type storeFixture struct {
	operation  tenantapi.OnboardingOperation
	lease      tenantapi.AttemptLease
	transition tenantapi.OnboardingTransition
	claim      tenantapi.ClaimedActivation
	complete   int
	err        error
}

func (store *storeFixture) Accept(context.Context, tenantapi.OnboardingPlan) (
	tenantapi.OnboardingOperation, error,
) {
	return store.operation, store.err
}
func (store *storeFixture) Claim(context.Context, tenantapi.ClaimAttemptRequest) (
	tenantapi.AttemptLease, error,
) {
	return store.lease, store.err
}
func (store *storeFixture) RecordReceipt(context.Context, tenantapi.OnboardingReceipt, int64) (
	tenantapi.OnboardingTransition, error,
) {
	return store.transition, store.err
}
func (store *storeFixture) ReceiveAuthorityAck(
	context.Context, tenantapi.AuthorityAcknowledgement, int64,
) (tenantapi.OnboardingTransition, error) {
	return store.transition, store.err
}
func (store *storeFixture) ClaimActivation(context.Context, ActivationRequest) (
	tenantapi.ClaimedActivation, error,
) {
	return store.claim, store.err
}
func (store *storeFixture) CompleteActivation(context.Context, ActivationCompletion) error {
	store.complete++
	return store.err
}
func (store *storeFixture) Reconcile(context.Context, ReconcileRequest) (
	tenantapi.OnboardingTransition, error,
) {
	return store.transition, store.err
}

type activationTransactionFixture struct {
	storeFixture
	boundSchema string
	boundOwners []ActivationOwner
}

func (transaction *activationTransactionFixture) bindTenantSchema(
	_ context.Context,
	schema string,
) error {
	transaction.boundSchema = schema
	return transaction.err
}

func (transaction *activationTransactionFixture) bindOwner(
	owner ActivationOwner,
) (ActivationOwnerExecutor, error) {
	transaction.boundOwners = append(transaction.boundOwners, owner)
	return ActivationOwnerExecutor{}, transaction.err
}

func (transaction *activationTransactionFixture) Activate(
	ctx context.Context,
	_ ActivationRequest,
	completion ActivationCompletion,
	callback ActivationTransactionCallback,
) error {
	if transaction.err != nil {
		return transaction.err
	}
	resources := ActivationResources{BindTenantSchema: transaction.bindTenantSchema,
		BindOwner: transaction.bindOwner}
	if err := callback(transaction.claim, resources); err != nil {
		return err
	}
	completion.ExpectedRevision = transaction.claim.Revision
	return transaction.CompleteActivation(ctx, completion)
}

type unitFixture struct {
	calls       int
	transaction *activationTransactionFixture
}

func (unit *unitFixture) Run(
	ctx context.Context,
	callback func(context.Context, ActivationTransaction) error,
) error {
	unit.calls++
	return callback(ctx, unit.transaction)
}

func planFixture(t *testing.T) tenantapi.OnboardingPlan {
	t.Helper()
	tenantID := uuid.New()
	schema, _ := domain.SchemaNameFromUUID(tenantID.String())
	now := time.Now().UTC().Truncate(time.Microsecond)
	input := tenantapi.OnboardingPlanInput{OperationID: uuid.New(), EnvironmentID: uuid.New(),
		TenantID: tenantID, FirstUserID: uuid.New(), RootOrganizationID: uuid.New(),
		ProtectedAdminBindingID: uuid.New(), InstanceOwnerBindingID: uuid.New(),
		Kind:           tenantapi.OnboardingEnvironmentInitialize,
		DeploymentMode: tenantapi.DeploymentModeOnPremise,
		BindingPurpose: tenantapi.TenantBindingOnPremisePrimary,
		SchemaName:     schema.String(), IdempotencyKey: "initialize",
		RequestSHA256: strings.Repeat("1", 64), ManifestSHA256: strings.Repeat("2", 64),
		Offering: tenantapi.BaseOfferingPin{VersionID: uuid.New(), Revision: 1,
			SHA256: strings.Repeat("3", 64), EffectiveAt: now,
			Source: tenantapi.OfferingSourcePin{Sequence: 1,
				SHA256: strings.Repeat("4", 64), Present: true}},
		RecoverySigning: tenantapi.RecoverySigningPin{Generation: 1,
			TrustBundleSHA256: strings.Repeat("5", 64), Present: true},
		Authority: tenantapi.ProtectedAuthorityPlan{RegistryBindingID: uuid.New(),
			ScopeKind: tenantapi.AuthorityScopeEnvironment}, CreatedAt: now}
	input.Authority.ScopeID = input.EnvironmentID
	plan, err := tenantapi.NewOnboardingPlan(input)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func coordinatorFixture(
	t *testing.T,
	plan tenantapi.OnboardingPlan,
) (*Coordinator, *storeFixture, *activationTransactionFixture) {
	t.Helper()
	operation, _ := tenantapi.NewOnboardingOperation(
		plan, tenantapi.OnboardingAccepted, 1, plan.CreatedAt(),
	)
	store := &storeFixture{operation: operation}
	transaction := &activationTransactionFixture{}
	coordinator, err := New(store, &unitFixture{transaction: transaction})
	if err != nil {
		t.Fatal(err)
	}
	return coordinator, store, transaction
}
