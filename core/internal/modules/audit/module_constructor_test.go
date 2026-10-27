package audit

import (
	"context"
	"net/http"
	"testing"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type constructorIDs struct{}

func (constructorIDs) NewUUIDv7() (uuid.UUID, error) {
	return uuid.MustParse("019c0000-0000-7000-8000-000000000501"), nil
}

type constructorClock struct{}

func (constructorClock) Now() time.Time {
	return time.Date(2026, 8, 2, 12, 0, 0, 0, time.UTC)
}

type constructorGenerations struct{}

func (constructorGenerations) ResolveTenantAuditGenerations(
	context.Context, uuid.UUID,
) (int64, int64, error) {
	return 7, 11, nil
}

type constructorScopes struct{}

func (constructorScopes) InTenant(
	context.Context, platformtx.ScopeRequest,
	func(context.Context, platformtx.TenantScope) error,
) error {
	return nil
}

type constructorAccess struct{}

func (constructorAccess) Verify(
	_ context.Context, request HTTPAccessRequest,
) (HTTPAccess, error) {
	return NewHTTPAccess(
		uuid.MustParse("019c0000-0000-7000-8000-000000000502"),
		uuid.MustParse("019c0000-0000-7000-8000-000000000503"),
		2, 3, 4, request.Permission, request.Operation)
}

func TestModuleConstructorsProduceReadyFiveRouteRuntime(t *testing.T) {
	if _, err := NewApplication(nil, constructorClock{}); err == nil {
		t.Fatal("nil ID source was accepted")
	}
	core, err := NewApplication(constructorIDs{}, constructorClock{})
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := NewHTTPRuntime(constructorGenerations{}, constructorScopes{}, core)
	if err != nil || !runtime.Ready() {
		t.Fatalf("runtime ready=%v err=%v", runtime != nil && runtime.Ready(), err)
	}
	handler, err := NewHTTPHandler(HTTPHandlerDependencies{
		Access: constructorAccess{}, Application: runtime, Report: func(error) {},
		Correlation: func(*http.Request) string { return "audit-constructor" },
	})
	routes := 0
	if handler != nil {
		routes = len(handler.Routes())
	}
	if err != nil || handler == nil || !handler.Ready() || routes != 5 {
		t.Fatalf("handler ready=%v routes=%d err=%v",
			handler != nil && handler.Ready(), routes, err)
	}
}
