package customerlifecycle

import (
	"context"
	"errors"
	"testing"
	"time"

	lifecycle "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerlifecycle"
	"github.com/google/uuid"
)

var (
	testEnvironment = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	testControl     = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	testTarget      = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	testActor       = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	testOperation   = uuid.MustParse("55555555-5555-4555-8555-555555555555")
	testNow         = time.Date(2026, 8, 2, 3, 23, 0, 0, time.UTC)
)

type fixtureRow struct{ scan func(...any) error }

func (row fixtureRow) Scan(targets ...any) error { return row.scan(targets...) }

type queryCapture struct {
	statement string
	arguments []any
	row       Row
}

func (capture *queryCapture) query(_ context.Context, statement string, arguments ...any) Row {
	capture.statement, capture.arguments = statement, arguments
	return capture.row
}

func newStore(t *testing.T, capture *queryCapture) *Store {
	t.Helper()
	store, err := NewStore(capture.query, func(err error) bool { return errors.Is(err, errMissing) })
	if err != nil {
		t.Fatal(err)
	}
	return store
}

var errMissing = errors.New("missing")

func TestNewStoreRejectsMissingQuery(t *testing.T) {
	if _, err := NewStore(nil, func(error) bool { return false }); err == nil {
		t.Fatal("missing query accepted")
	}
}

func TestListDecodesBoundedDocument(t *testing.T) {
	document := `[{"id":"33333333-3333-4333-8333-333333333333","name":"A","slug":"a","status":"ACTIVE","revision":2,"settings_revision":3,"created_at":"2026-08-02T03:23:00Z","updated_at":"2026-08-02T03:23:00Z","offboarded_at":null,"purged_at":null}]`
	capture := &queryCapture{row: fixtureRow{scan: func(targets ...any) error {
		*targets[0].(*[]byte) = []byte(document)
		*targets[1].(**uuid.UUID) = nil
		return nil
	}}}
	page, err := newStore(t, capture).List(context.Background(), lifecycle.ListStoreRequest{
		EnvironmentID: testEnvironment, Limit: 10,
	})
	if err != nil || len(page.Items) != 1 || page.Items[0].Revision != 2 ||
		len(capture.arguments) != 3 || capture.statement != listSQL {
		t.Fatalf("page=%+v err=%v args=%+v", page, err, capture.arguments)
	}
}

func TestMutateScansOutcomeAndTenant(t *testing.T) {
	capture := &queryCapture{row: tenantMutationRow("applied", false)}
	result, err := newStore(t, capture).Mutate(context.Background(), lifecycle.MutationStoreRequest{
		EnvironmentID: testEnvironment, ControlTenantID: testControl, ActorID: testActor,
		TargetTenantID: testTarget, Kind: lifecycle.MutationSuspend, ExpectedRevision: 1,
		IdempotencyKey: "tenant-suspend-key-01", RequestSHA256: [32]byte{1}, Now: testNow,
	})
	if err != nil || result.Outcome != "applied" || result.Tenant.ID != testTarget ||
		result.Tenant.CreatedAt.Location() != time.UTC || len(capture.arguments) != 10 {
		t.Fatalf("result=%+v err=%v args=%d", result, err, len(capture.arguments))
	}
}

func tenantMutationRow(outcome string, replay bool) Row {
	return fixtureRow{scan: func(targets ...any) error {
		*targets[0].(*string), *targets[1].(*bool) = outcome, replay
		id, name, slug, status := testTarget, "Tenant", "tenant", "SUSPENDED"
		revision, settingsRevision := int64(2), int64(1)
		createdAt, updatedAt := testNow, testNow
		*targets[2].(**uuid.UUID), *targets[3].(**string) = &id, &name
		*targets[4].(**string), *targets[5].(**string) = &slug, &status
		*targets[6].(**int64), *targets[7].(**int64) = &revision, &settingsRevision
		*targets[8].(**time.Time), *targets[9].(**time.Time) = &createdAt, &updatedAt
		*targets[10].(**time.Time), *targets[11].(**time.Time) = nil, nil
		return nil
	}}
}

func TestCompleteOffboardUsesOnlyThreeParameters(t *testing.T) {
	capture := &queryCapture{row: operationRow()}
	store := newStore(t, capture)
	_, err := store.CompleteOffboard(context.Background(), lifecycle.ClaimedOperation{
		Operation: lifecycle.Operation{ID: testOperation}, WorkerID: testActor,
	}, testNow)
	if err != nil || capture.statement != completeOffboardSQL || len(capture.arguments) != 3 {
		t.Fatalf("err=%v statement=%q args=%+v", err, capture.statement, capture.arguments)
	}
}

func operationRow() Row {
	return fixtureRow{scan: func(targets ...any) error {
		*targets[0].(*uuid.UUID), *targets[1].(*uuid.UUID) = testOperation, testControl
		*targets[2].(*uuid.UUID), *targets[3].(*string) = testTarget, lifecycle.OperationOffboard
		*targets[4].(*string), *targets[5].(*int64) = "SUCCEEDED", 2
		*targets[6].(*bool) = false
		*targets[7].(*time.Time), *targets[8].(*time.Time) = testNow, testNow
		completed := testNow
		*targets[9].(**time.Time), *targets[10].(**string) = &completed, nil
		return nil
	}}
}

func TestMissingRowsMapToNotFound(t *testing.T) {
	capture := &queryCapture{row: fixtureRow{scan: func(...any) error { return errMissing }}}
	_, err := newStore(t, capture).Get(context.Background(), testEnvironment, testTarget)
	if !errors.Is(err, lifecycle.ErrNotFound) {
		t.Fatalf("err=%v", err)
	}
}
