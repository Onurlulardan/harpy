package compatibility

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

var (
	compatNow = time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
	runID     = uuid.MustParse("018f47f2-4a7c-7000-8000-000000000001")
	tenantID  = uuid.MustParse("018f47f2-4a7c-7000-8000-000000000002")
	legacyID  = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	queryID   = uuid.MustParse("018f47f2-4a7c-7000-8000-000000000003")
	versionID = uuid.MustParse("018f47f2-4a7c-7000-8000-000000000004")
	releaseID = uuid.MustParse("018f47f2-4a7c-7000-8000-000000000005")
	authority = AdoptionAuthority{AuthorityID: uuid.MustParse(
		"018f47f2-4a7c-7000-8000-000000000006"), TenantID: tenantID, SchemaName: "t_fixture"}
)

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }

type adoptionStub struct {
	item      AdoptionItem
	receipt   Acceptance
	recordErr error
	acceptErr error
}

func (stub *adoptionStub) BackfillItem(
	_ context.Context, _ AdoptionAuthority, item AdoptionItem,
) (AdoptionItem, bool, error) {
	stub.item = item
	return item, false, stub.recordErr
}

func (stub *adoptionStub) Accept(
	_ context.Context, _ AdoptionAuthority, receipt Acceptance,
) (Acceptance, bool, error) {
	stub.receipt = receipt
	return receipt, false, stub.acceptErr
}

type legacyStub struct{ mapping LegacyMapping }

func (stub *legacyStub) Resolve(_ context.Context, _ LegacyKey) (LegacyMapping, error) {
	return stub.mapping, nil
}
func (stub *legacyStub) RecordHit(
	_ context.Context, value LegacyMapping, at time.Time,
) (LegacyMapping, error) {
	value.HitCount++
	value.LastHitAt = &at
	stub.mapping = value
	return value, nil
}
func (stub *legacyStub) FinalizeConsumers(
	_ context.Context, value LegacyMapping, _ [32]byte, at time.Time,
) (LegacyMapping, error) {
	value.ConsumersAckedAt = &at
	stub.mapping = value
	return value, nil
}
func (stub *legacyStub) Retire(
	_ context.Context, value LegacyMapping, _ RetirementEvidence, _ time.Time,
) (LegacyMapping, error) {
	value.State = "retired"
	stub.mapping = value
	return value, nil
}

type authorityStub struct{ evidence RetirementEvidence }

func (stub authorityStub) Evidence(context.Context) (RetirementEvidence, error) {
	return stub.evidence, nil
}

func applicationFixture(t *testing.T, mapping LegacyMapping) (*Application, *adoptionStub) {
	t.Helper()
	adoption := &adoptionStub{}
	application, err := NewApplication(adoption, &legacyStub{mapping: mapping},
		authorityStub{evidence: RetirementEvidence{AllTenantsV2: true,
			SHA256: [32]byte{9}, ObservedAt: compatNow}}, fixedClock{now: compatNow})
	if err != nil {
		t.Fatal(err)
	}
	return application, adoption
}

func mappingFixture() LegacyMapping {
	lastHit := compatNow.Add(-31 * 24 * time.Hour)
	acked := compatNow.Add(-30 * 24 * time.Hour)
	return LegacyMapping{Key: LegacyKey{TenantIdentity: tenantID.String(), Method: "GET",
		Path: "/api/openapi/orders", Slug: "orders"}, TenantID: tenantID,
		LegacyEndpointID: legacyID, ReleaseID: releaseID, PublicID: [16]byte{1},
		State: "active", HitCount: 4, LastHitAt: &lastHit, ConsumersAckedAt: &acked,
		RemoveAfter: compatNow.Add(-time.Hour)}
}
