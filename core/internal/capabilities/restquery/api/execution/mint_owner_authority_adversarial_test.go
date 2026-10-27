package execution

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"unsafe"

	restqueryapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestMintRejectsCatalogOwnerOutsideScopeBeforePort(t *testing.T) {
	ctx, scope := opaqueMintScope(t, "forms")
	reference := mintCatalogReference(t, "crm")
	var closureCalls, portCalls atomic.Int32
	program, err := MintProgram(ctx, scope, reference, func(
		context.Context, SnapshotDriver,
	) error {
		closureCalls.Add(1)
		return nil
	})
	if err == nil {
		registry := newTestRegistry(t, "crm", testPort(func(
			transaction.TenantScope, Program,
		) error {
			portCalls.Add(1)
			return nil
		}))
		_ = registry.Run(program)
	}
	if !errors.Is(err, errExecutionContract) ||
		closureCalls.Load() != 0 || portCalls.Load() != 0 {
		t.Fatalf("error=%v closure=%d port=%d", err,
			closureCalls.Load(), portCalls.Load())
	}
}

func TestMintAcceptsCatalogOwnerInsideExactScope(t *testing.T) {
	ctx, scope := opaqueMintScope(t, "forms")
	var closureCalls atomic.Int32
	program, err := MintProgram(ctx, scope, mintCatalogReference(t, "forms"), func(
		context.Context, SnapshotDriver,
	) error {
		closureCalls.Add(1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	registry := newTestRegistry(t, "forms", testPort(func(
		_ transaction.TenantScope,
		received Program,
	) error {
		return received.Run(&testDriver{owner: "forms", scope: scope})
	}))
	if err := registry.Run(program); err != nil || closureCalls.Load() != 1 {
		t.Fatalf("error=%v closure=%d", err, closureCalls.Load())
	}
}

type opaqueMintContext struct {
	context.Context
	state any
}

func (ctx opaqueMintContext) Value(key any) any {
	typeOfKey := reflect.TypeOf(key)
	if typeOfKey != nil && typeOfKey.PkgPath() == reflect.TypeOf(
		transaction.TenantScope{},
	).PkgPath() && typeOfKey.Name() == "scopeContextKey" {
		return ctx.state
	}
	return ctx.Context.Value(key)
}

// TenantScope intentionally has no public constructor. This test creates its
// actual private state by reflection, then writes only named fields required by
// the public scope contract. No production path receives this fixture.
func opaqueMintScope(
	t *testing.T,
	owner string,
) (context.Context, transaction.TenantScope) {
	t.Helper()
	operation, err := transaction.NewSnapshotAuthorityIssuer().Issue(
		snapshotOperationID, map[string]int64{owner: 1},
	)
	if err != nil {
		t.Fatal(err)
	}
	var scope transaction.TenantScope
	scopeType := reflect.TypeOf(scope)
	stateType := scopeType.Field(0).Type.Elem()
	state := reflect.New(stateType)
	setOpaqueField(t, state.Elem(), "actorID", reflect.ValueOf(uuid.MustParse(
		"22222222-2222-4222-8222-222222222222",
	)))
	setOpaqueField(t, state.Elem(), "tenantID", reflect.ValueOf(uuid.MustParse(
		"11111111-1111-4111-8111-111111111111",
	)))
	setOpaqueField(t, state.Elem(), "operation", reflect.ValueOf(operation))
	setOpaqueField(t, state.Elem(), "mode", reflect.ValueOf(transaction.SnapshotQuery))
	active := state.Elem().FieldByName("active")
	(*atomic.Bool)(unsafe.Pointer(active.UnsafeAddr())).Store(true)
	scopeField := reflect.ValueOf(&scope).Elem().Field(0)
	setOpaqueValue(t, scopeField, state)
	ctx := opaqueMintContext{Context: context.Background(), state: state.Interface()}
	if transaction.NetworkAllowed(ctx) || !transaction.ContextOwnsScope(ctx, scope) {
		t.Fatal("opaque scope fixture did not preserve context ownership")
	}
	return ctx, scope
}

func setOpaqueField(t *testing.T, target reflect.Value, name string, value reflect.Value) {
	t.Helper()
	field := target.FieldByName(name)
	if !field.IsValid() {
		t.Fatalf("scope state field %q is unavailable", name)
	}
	setOpaqueValue(t, field, value)
}

func setOpaqueValue(t *testing.T, target, value reflect.Value) {
	t.Helper()
	if !value.Type().AssignableTo(target.Type()) || !target.CanAddr() {
		t.Fatal("opaque scope fixture type mismatch")
	}
	reflect.NewAt(target.Type(), unsafe.Pointer(target.UnsafeAddr())).Elem().Set(value)
}

func mintCatalogReference(t *testing.T, owner string) restqueryapi.CatalogReference {
	t.Helper()
	hash, err := restqueryapi.NewCatalogHash(strings.Repeat("ab", 32))
	if err != nil {
		t.Fatal(err)
	}
	reference, err := restqueryapi.NewCatalogReference(
		owner,
		"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		hash,
	)
	if err != nil {
		t.Fatal(err)
	}
	return reference
}
