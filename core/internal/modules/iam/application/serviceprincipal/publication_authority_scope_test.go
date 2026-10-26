package serviceprincipal

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"
	"unsafe"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type authorityScopeContext struct {
	context.Context
	state any
}

func (ctx authorityScopeContext) Value(key any) any {
	keyType := reflect.TypeOf(key)
	scopeType := reflect.TypeOf(platformtx.TenantScope{})
	if keyType != nil && keyType.PkgPath() == scopeType.PkgPath() &&
		keyType.Name() == "scopeContextKey" {
		return ctx.state
	}
	return ctx.Context.Value(key)
}

func publicationAuthorityScope(
	t *testing.T,
) (context.Context, platformtx.TenantScope, uuid.UUID, uuid.UUID) {
	t.Helper()
	actorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	tenantID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	operation, err := platformtx.NewAuthorityIssuer().IssueMultiGroup(
		publicationAuthorityOperation,
		[]platformtx.CommandAuthorityGroup{
			{ID: "iam_system_resources", Generation: 5},
			{ID: "workflow", Generation: 7},
		},
		[]platformtx.CommandAuthorityDomain{
			{Name: "iam", Generation: 5},
			{Name: "system_resource_layout", Generation: 5},
			{Name: "workflow", Generation: 7},
		},
		[]string{"iam", "system_resource_layout", "workflow"},
	)
	if err != nil {
		t.Fatal(err)
	}
	var scope platformtx.TenantScope
	stateType := reflect.TypeOf(scope).Field(0).Type.Elem()
	state := reflect.New(stateType)
	setAuthorityScopeField(t, state.Elem(), "actorID", reflect.ValueOf(actorID))
	setAuthorityScopeField(t, state.Elem(), "tenantID", reflect.ValueOf(tenantID))
	setAuthorityScopeField(t, state.Elem(), "operation", reflect.ValueOf(operation))
	setAuthorityScopeField(
		t, state.Elem(), "mode", reflect.ValueOf(platformtx.Command),
	)
	active := state.Elem().FieldByName("active")
	(*atomic.Bool)(unsafe.Pointer(active.UnsafeAddr())).Store(true)
	setAuthorityScopeValue(t, reflect.ValueOf(&scope).Elem().Field(0), state)
	ctx := authorityScopeContext{Context: context.Background(), state: state.Interface()}
	if !platformtx.ContextOwnsScope(ctx, scope) {
		t.Fatal("publication authority scope is not owned")
	}
	return ctx, scope, actorID, tenantID
}

func setAuthorityScopeField(
	t *testing.T,
	target reflect.Value,
	name string,
	value reflect.Value,
) {
	t.Helper()
	field := target.FieldByName(name)
	if !field.IsValid() {
		t.Fatalf("scope field %q is unavailable", name)
	}
	setAuthorityScopeValue(t, field, value)
}

func setAuthorityScopeValue(t *testing.T, target, value reflect.Value) {
	t.Helper()
	if !value.Type().AssignableTo(target.Type()) || !target.CanAddr() {
		t.Fatal("scope fixture type mismatch")
	}
	reflect.NewAt(target.Type(), unsafe.Pointer(target.UnsafeAddr())).Elem().Set(value)
}
