package audit

import (
	"context"
	"reflect"
	"testing"

	audithttp "github.com/Onurlulardan/harpy/core/internal/modules/audit/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type runtimeGenerationProbe struct {
	calls    int
	tenantID uuid.UUID
}

func (probe *runtimeGenerationProbe) ResolveTenantAuditGenerations(
	_ context.Context, tenantID uuid.UUID,
) (int64, int64, error) {
	probe.calls++
	probe.tenantID = tenantID
	return 7, 11, nil
}

type runtimeScopeProbe struct {
	calls   int
	request platformtx.ScopeRequest
}

func (probe *runtimeScopeProbe) InTenant(
	_ context.Context, request platformtx.ScopeRequest,
	_ func(context.Context, platformtx.TenantScope) error,
) error {
	probe.calls++
	probe.request = request
	return nil
}

func TestHTTPRuntimePinsAuditSnapshotAuthority(t *testing.T) {
	actor := uuid.MustParse("019c0000-0000-7000-8000-000000000901")
	tenant := uuid.MustParse("019c0000-0000-7000-8000-000000000902")
	access, err := NewHTTPAccess(actor, tenant, 2, 3, 4,
		PermissionRead, application.OperationList)
	if err != nil {
		t.Fatal(err)
	}
	core, err := NewApplication(constructorIDs{}, constructorClock{})
	if err != nil {
		t.Fatal(err)
	}
	generations, scopes := &runtimeGenerationProbe{}, &runtimeScopeProbe{}
	runtime, err := NewHTTPRuntime(generations, scopes, core)
	if err != nil {
		t.Fatal(err)
	}
	_, err = runtime.List(context.Background(), access,
		audithttp.Filter{Page: 1, Limit: 20}, application.OperationList)
	if err != nil || generations.calls != 1 || scopes.calls != 1 ||
		generations.tenantID != tenant || scopes.request.Mode != platformtx.SnapshotQuery {
		t.Fatalf("generation=%d scope=%d tenant=%s mode=%s err=%v",
			generations.calls, scopes.calls, generations.tenantID, scopes.request.Mode, err)
	}
	operation := reflect.ValueOf(scopes.request.Operation)
	if operation.FieldByName("operationID").String() != application.OperationList {
		t.Fatal("snapshot operation ID drifted")
	}
	domains := operation.FieldByName("domains")
	if domains.Len() != 1 || domains.Index(0).FieldByName("name").String() != "audit" ||
		domains.Index(0).FieldByName("generation").Int() != 11 {
		t.Fatal("snapshot audit generation drifted")
	}
}
