package resourceentitlement

import (
	"context"
	"testing"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	"github.com/google/uuid"
)

var resourceNow = time.Date(2026, time.August, 2, 10, 0, 0, 123000, time.UTC)

type fakeResourceStore struct {
	inspection          Inspection
	reconciliation      ReconciliationInspection
	feature             bool
	inspectErr          error
	reconcileInspectErr error
	featureErr          error
	recordErr           error
	reconcileRecordErr  error
	records             []MutationRecord
	reconcileRecords    []ReconciliationRecord
	inspectCalls        int
}

func (store *fakeResourceStore) Inspect(
	_ context.Context,
	_ billingapi.MeteredResourceMutationRequest,
) (Inspection, error) {
	store.inspectCalls++
	return store.inspection, store.inspectErr
}

func (store *fakeResourceStore) InspectReconciliation(
	_ context.Context,
	_ billingapi.MeteredResourceReconcileRequest,
) (ReconciliationInspection, error) {
	store.inspectCalls++
	return store.reconciliation, store.reconcileInspectErr
}

func (store *fakeResourceStore) ResolveFeature(
	_ context.Context,
	_ entitlement.Key,
) (bool, error) {
	return store.feature, store.featureErr
}

func (store *fakeResourceStore) Record(_ context.Context, record MutationRecord) error {
	if store.recordErr == nil {
		store.records = append(store.records, record)
	}
	return store.recordErr
}

func (store *fakeResourceStore) RecordReconciliation(
	_ context.Context,
	record ReconciliationRecord,
) error {
	if store.reconcileRecordErr == nil {
		store.reconcileRecords = append(store.reconcileRecords, record)
	}
	return store.reconcileRecordErr
}

type fakeUsage struct {
	snapshot    GaugeSnapshot
	inspectErr  error
	adjustErr   error
	allowed     bool
	adjustCalls []GaugeAdjustmentRequest
}

func (usage *fakeUsage) InspectGauge(
	context.Context,
	entitlement.Key,
) (GaugeSnapshot, error) {
	return usage.snapshot, usage.inspectErr
}

func (usage *fakeUsage) AdjustGauge(
	_ context.Context,
	request GaugeAdjustmentRequest,
) (GaugeAdjustment, error) {
	usage.adjustCalls = append(usage.adjustCalls, request)
	if usage.adjustErr != nil {
		return GaugeAdjustment{}, usage.adjustErr
	}
	result := GaugeAdjustment{
		DecisionID: mustV7(), Allowed: usage.allowed,
		Before: usage.snapshot.Current, After: usage.snapshot.Current,
		Limit: usage.snapshot.Limit,
	}
	if usage.allowed {
		result.After += request.Delta
		result.Entry = GaugeEntry{
			ID: mustV7(), DecisionID: result.DecisionID,
			EntitlementKey: request.EntitlementKey, PeriodID: request.PeriodID,
			OwnerFactID: request.OwnerFactID, Delta: request.Delta,
			Before: result.Before, After: result.After,
		}
	}
	return result, nil
}

func resourceFixture(t *testing.T, current, limit int64) (*fakeResourceStore, *fakeUsage) {
	t.Helper()
	key, err := entitlement.NewKey("api.active_endpoints")
	if err != nil {
		t.Fatal(err)
	}
	definition, err := entitlement.NewDefinition(
		key, entitlement.Integer, entitlement.Gauge, entitlement.Never,
		entitlement.Hard, true,
	)
	if err != nil {
		t.Fatal(err)
	}
	limitValue, err := entitlement.IntegerValue(limit)
	if err != nil {
		t.Fatal(err)
	}
	store := &fakeResourceStore{feature: true}
	usage := &fakeUsage{allowed: current < limit, snapshot: GaugeSnapshot{
		Definition: definition, Limit: limitValue, PeriodID: mustV7(),
		Current: current,
	}}
	return store, usage
}

func mutationRequest(mutation billingapi.MeteredResourceMutation) billingapi.MeteredResourceMutationRequest {
	return billingapi.MeteredResourceMutationRequest{
		Resource: billingapi.MeteredResourceAPIEndpoint, Mutation: mutation,
		TenantID: mustV7(), ActorID: mustV7(), ProposedOperationID: mustV7(),
		ProposedOwnerFactID: mustV7(), IdempotencyKey: "resource-1",
		RequestSHA256: [32]byte{1},
	}
}

func mustV7() uuid.UUID { return uuid.Must(uuid.NewV7()) }
