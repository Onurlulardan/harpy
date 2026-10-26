package support

import (
	"context"
	"crypto/sha256"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

type runtimeAccessFixture struct{}

func (runtimeAccessFixture) Verify(_ context.Context, request HTTPAccessRequest) (HTTPAccess, error) {
	tenantID, slug, name := uuid.Nil, "", ""
	if request.Scope != HTTPPlatformScope {
		tenantID = uuid.MustParse("10000000-0000-4000-8000-000000000001")
		slug, name = "tenant-a", "Tenant A"
	}
	return NewHTTPAccess(uuid.MustParse("20000000-0000-4000-8000-000000000001"),
		tenantID, 1, 1, 1, request.Scope, request.Permission, request.Operation,
		"Ada", "ada@example.com", slug, name)
}

type runtimeRateFixture struct{}

func (runtimeRateFixture) Resolve(*http.Request) ([32]byte, error) {
	return sha256.Sum256([]byte("support-runtime-rate")), nil
}

type runtimeClockFixture struct{}

func (runtimeClockFixture) Now() time.Time {
	return time.Date(2026, 8, 2, 12, 0, 0, 123456000, time.UTC)
}

type runtimeIDFixture struct{}

func (runtimeIDFixture) NewV7() (RuntimeID, error) {
	return RuntimeID(uuid.MustParse("01900000-0000-7000-8000-000000000001")), nil
}

func TestRuntimeFacadeAssemblesAllSeventeenRoutes(t *testing.T) {
	runner := func(context.Context, RuntimeTransactionMode,
		func(RuntimeDatabase) error,
	) error {
		return errors.New("transaction must not run during assembly")
	}
	handler, err := NewRuntimeHTTPHandler(RuntimeHTTPDependencies{Run: runner,
		RowNotFound: func(error) bool { return false }, Clock: runtimeClockFixture{},
		IDs: runtimeIDFixture{}, HTTP: HTTPHandlerDependencies{Access: runtimeAccessFixture{},
			RateKeys: runtimeRateFixture{}, Booking: HTTPBookingConfig{}, Report: func(error) {}}})
	if err != nil {
		t.Fatal(err)
	}
	if !handler.Ready() || len(handler.Routes()) != 17 {
		t.Fatalf("support runtime ready=%v routes=%d", handler.Ready(), len(handler.Routes()))
	}
}
