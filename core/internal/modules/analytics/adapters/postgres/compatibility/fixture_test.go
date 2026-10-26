package compatibility

import (
	"context"
	"testing"
	"time"

	compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"
	"github.com/google/uuid"
)

type scanRow func([]any) error

func (row scanRow) Scan(targets ...any) error { return row(targets) }

type databaseStub struct {
	sql  string
	args []any
	row  row
}

func (stub *databaseStub) QueryRow(_ context.Context, sql string, args ...any) row {
	stub.sql = sql
	stub.args = args
	return stub.row
}

var (
	testTime      = time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	testRun       = uuid.MustParse("018f47f2-4a7c-7000-8000-000000000011")
	testTenant    = uuid.MustParse("018f47f2-4a7c-7000-8000-000000000012")
	testLegacy    = uuid.MustParse("11111111-1111-4111-8111-111111111112")
	testQuery     = uuid.MustParse("018f47f2-4a7c-7000-8000-000000000013")
	testVersion   = uuid.MustParse("018f47f2-4a7c-7000-8000-000000000014")
	testRelease   = uuid.MustParse("018f47f2-4a7c-7000-8000-000000000015")
	testAuthority = compatibilityapp.AdoptionAuthority{AuthorityID: uuid.MustParse(
		"018f47f2-4a7c-7000-8000-000000000016"), TenantID: testTenant, SchemaName: "t_fixture"}
)

func newStore(t *testing.T, row row) (*Store, *databaseStub) {
	t.Helper()
	database := &databaseStub{row: row}
	store, err := NewStore(database)
	if err != nil {
		t.Fatal(err)
	}
	return store, database
}

func legacyMapping() compatibilityapp.LegacyMapping {
	last := testTime.Add(-31 * 24 * time.Hour)
	acked := testTime.Add(-30 * 24 * time.Hour)
	return compatibilityapp.LegacyMapping{Key: compatibilityapp.LegacyKey{
		TenantIdentity: testTenant.String(), Method: "GET",
		Path: "/api/openapi/orders", Slug: "orders"}, TenantID: testTenant,
		LegacyEndpointID: testLegacy, ReleaseID: testRelease, PublicID: [16]byte{1},
		State: "active", HitCount: 3, LastHitAt: &last, ConsumersAckedAt: &acked,
		RemoveAfter: testTime.Add(-time.Hour)}
}

func mappingRow(value compatibilityapp.LegacyMapping) row {
	return scanRow(func(targets []any) error {
		*targets[0].(*string) = value.Key.TenantIdentity
		*targets[1].(*string) = value.Key.Method
		*targets[2].(*string) = value.Key.Path
		*targets[3].(*string) = value.Key.Slug
		*targets[4].(*uuid.UUID) = value.TenantID
		*targets[5].(*uuid.UUID) = value.LegacyEndpointID
		*targets[6].(*[]byte) = append([]byte(nil), value.PublicID[:]...)
		*targets[7].(*uuid.UUID) = value.ReleaseID
		*targets[8].(*string) = value.State
		*targets[9].(*uint64) = value.HitCount
		*targets[10].(**time.Time) = value.LastHitAt
		*targets[11].(**time.Time) = value.ConsumersAckedAt
		*targets[12].(*time.Time) = value.RemoveAfter
		return nil
	})
}
