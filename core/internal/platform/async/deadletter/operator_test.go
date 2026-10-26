package deadletter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestOperatorRequiresExplicitTenantOrPlatformPermission(t *testing.T) {
	tenantA, tenantB := uuid.New(), uuid.New()
	store := &operatorStoreFake{}
	permissions := &permissionFake{}
	operator, err := NewOperator(permissions, store, fixedIDs{value: uuid.New()})
	if err != nil {
		t.Fatal(err)
	}
	actor := uuid.New()
	if _, err := operator.List(context.Background(), ListRequest{
		ActorID: actor, TenantIDs: []uuid.UUID{tenantA, tenantB}, Limit: 10,
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("tenant cross-scope error = %v", err)
	}
	if permissions.calls != 0 || store.listCalls != 0 {
		t.Fatal("invalid tenant scope reached permission or storage")
	}
	if _, err := operator.List(context.Background(), ListRequest{
		ActorID: actor, TenantIDs: []uuid.UUID{tenantA, tenantB}, Platform: true, Limit: 10,
	}); err != nil {
		t.Fatal(err)
	}
	if permissions.last.Permission != PermissionRead || !permissions.last.Platform ||
		len(store.tenants) != 2 {
		t.Fatal("platform scope was not explicit end to end")
	}
}

func TestOperatorDenialNeverTouchesStore(t *testing.T) {
	store := &operatorStoreFake{}
	permissions := &permissionFake{err: errors.New("denied")}
	operator, _ := NewOperator(permissions, store, fixedIDs{value: uuid.New()})
	_, err := operator.List(context.Background(), ListRequest{
		ActorID: uuid.New(), TenantIDs: []uuid.UUID{uuid.New()}, Limit: 1,
	})
	if !errors.Is(err, ErrUnauthorized) || store.listCalls != 0 {
		t.Fatalf("denied read reached storage: %v", err)
	}
}

func TestOperatorRetrySealsAuditAndServerGeneratedActionID(t *testing.T) {
	actionID := uuid.New()
	store := &operatorStoreFake{}
	permissions := &permissionFake{}
	operator, _ := NewOperator(permissions, store, fixedIDs{value: actionID})
	request := RetryRequest{
		ActorID: uuid.New(), TenantID: uuid.New(), RootID: uuid.New(),
		Reason: "consumer upgraded", Ticket: "OPS-42", CausationID: uuid.New(),
	}
	if _, err := operator.Retry(context.Background(), request); err != nil {
		t.Fatal(err)
	}
	if permissions.last.Permission != PermissionRetry || store.retry.ActionID() != actionID ||
		store.retry.Reason() != request.Reason || store.retry.Ticket() != request.Ticket {
		t.Fatal("retry audit or server-generated identity drifted")
	}
	request.Reason = " "
	if _, err := operator.Retry(context.Background(), request); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("blank audit reason error = %v", err)
	}
}

type fixedIDs struct{ value uuid.UUID }

func (source fixedIDs) NewUUID() uuid.UUID { return source.value }

type permissionFake struct {
	calls int
	last  PermissionCheck
	err   error
}

func (fake *permissionFake) Verify(_ context.Context, check PermissionCheck) error {
	fake.calls++
	fake.last = check
	return fake.err
}

type operatorStoreFake struct {
	listCalls int
	tenants   []uuid.UUID
	retry     AuthorizedRetry
}

func (fake *operatorStoreFake) List(_ context.Context, request AuthorizedList) (Page, error) {
	fake.listCalls++
	fake.tenants = request.TenantIDs()
	return Page{}, nil
}

func (fake *operatorStoreFake) Retry(_ context.Context, request AuthorizedRetry) (RetryResult, error) {
	fake.retry = request
	return RetryResult{ActionID: request.ActionID(), OccurredAt: time.Now()}, nil
}
