package application

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type globalCoordinatorStub struct {
	called  bool
	request platformtx.GlobalScopeRequest
}

func (stub *globalCoordinatorStub) InGlobal(
	_ context.Context,
	request platformtx.GlobalScopeRequest,
	_ func(context.Context, platformtx.GlobalScope) error,
) error {
	stub.called, stub.request = true, request
	return nil
}

func TestGlobalCatalogPublicSnapshotBuildsExactRequest(t *testing.T) {
	stub := &globalCoordinatorStub{}
	catalog, err := NewGlobalCatalog(stub)
	if err != nil {
		t.Fatal(err)
	}
	err = catalog.PublicSnapshot(
		context.Background(), platformtx.PublicBillingCatalogList, 7, 5,
		func(context.Context, platformtx.GlobalScope) error { return nil },
	)
	if err != nil || !stub.called {
		t.Fatalf("public global snapshot = %v, called=%v", err, stub.called)
	}
}

func TestGlobalCatalogRejectsInvalidCapabilityBeforeCoordinator(t *testing.T) {
	stub := &globalCoordinatorStub{}
	catalog, _ := NewGlobalCatalog(stub)
	err := catalog.Command(
		context.Background(), platformtx.GlobalCommandCapability{},
		func(
			context.Context, platformtx.GlobalScope, platformtx.OwnerAuthority,
		) error {
			return nil
		},
	)
	if !errors.Is(err, platformtx.ErrInvalidGlobalScopeRequest) || stub.called {
		t.Fatalf("invalid command = %v, called=%v", err, stub.called)
	}
	if _, err = NewGlobalCatalog(nil); err == nil {
		t.Fatal("nil coordinator accepted")
	}
}

func TestGlobalCatalogAuthenticatedSnapshotUsesReadRequest(t *testing.T) {
	stub := &globalCoordinatorStub{}
	catalog, _ := NewGlobalCatalog(stub)
	capability, err := platformtx.NewGlobalSnapshotCapabilityIssuer().Issue(
		uuid.New(), "billing.catalog.offering.list", "platform.billing.read",
		7, 5, 3, 2,
	)
	if err != nil {
		t.Fatal(err)
	}
	err = catalog.Snapshot(
		context.Background(), capability,
		func(context.Context, platformtx.GlobalScope) error { return nil },
	)
	if err != nil || !stub.called {
		t.Fatalf("authenticated snapshot = %v, called=%v", err, stub.called)
	}
}
