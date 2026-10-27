//go:build integration

package api

import (
	"context"
	"encoding/hex"
	"reflect"
	"strings"
	"testing"
	"time"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type catalogDemoProvisionReceipt struct {
	ID          string
	State       string
	Ready       bool
	RetryAt     time.Time
	CompletedAt time.Time
}

const integrationCatalogDemoTenantID = "76000000-0000-4000-8000-000000000071"

func catalogDemoProvision(
	t *testing.T,
	ctx context.Context,
	coordinator *platformtx.Coordinator,
	owner any,
	request any,
) catalogDemoProvisionReceipt {
	t.Helper()
	method := reflect.ValueOf(owner).MethodByName("Provision")
	if !method.IsValid() || method.Type().NumIn() != 3 ||
		reflect.TypeOf(request) != method.Type().In(2) {
		t.Fatal("Catalog component owner Provision ABI differs")
	}
	var receipt catalogDemoProvisionReceipt
	err := coordinator.InTenant(
		ctx, catalogDemoScopeRequest(t),
		func(scoped context.Context, scope platformtx.TenantScope) error {
			output := method.Call([]reflect.Value{
				reflect.ValueOf(scoped), reflect.ValueOf(scope), reflect.ValueOf(request),
			})
			if len(output) != 2 {
				t.Fatal("Catalog component owner Provision output ABI differs")
			}
			if !output[1].IsNil() {
				provisionErr, ok := output[1].Interface().(error)
				if !ok {
					t.Fatal("Catalog component owner Provision error ABI differs")
				}
				return provisionErr
			}
			receipt = catalogDemoReadReceipt(t, output[0])
			return nil
		},
	)
	if err != nil {
		t.Fatalf("provision Catalog component: %v", err)
	}
	return receipt
}

func catalogDemoReadReceipt(
	t *testing.T, value reflect.Value,
) catalogDemoProvisionReceipt {
	t.Helper()
	readString := func(name string) string {
		field := value.FieldByName(name)
		if !field.IsValid() || field.Kind() != reflect.String {
			t.Fatalf("Catalog provision receipt field %s differs", name)
		}
		return field.String()
	}
	readTime := func(name string) time.Time {
		field := value.FieldByName(name)
		if !field.IsValid() {
			t.Fatalf("Catalog provision receipt field %s differs", name)
		}
		result, ok := field.Interface().(time.Time)
		if !ok {
			t.Fatalf("Catalog provision receipt field %s is not time.Time", name)
		}
		return result
	}
	ready := value.FieldByName("Ready")
	if !ready.IsValid() || ready.Kind() != reflect.Bool {
		t.Fatal("Catalog provision receipt Ready field differs")
	}
	return catalogDemoProvisionReceipt{
		ID: readString("ID"), State: readString("State"), Ready: ready.Bool(),
		RetryAt: readTime("RetryAt"), CompletedAt: readTime("CompletedAt"),
	}
}

func catalogDemoScopeRequest(t *testing.T) platformtx.ScopeRequest {
	t.Helper()
	admit := reflect.ValueOf(platformtx.AdmitMultiGroupCommandScope)
	output := admit.Call([]reflect.Value{
		catalogDemoUUIDValue(t, admit.Type().In(0),
			"018f2000-0000-7000-8000-000000000010"),
		catalogDemoUUIDValue(t, admit.Type().In(1),
			integrationCatalogDemoTenantID),
		reflect.ValueOf("catalog.operation.resume"),
		reflect.ValueOf(
			[]platformtx.CommandAuthorityGroup{
				{ID: "analytics", Generation: 7},
				{ID: "billing_forms", Generation: 7},
				{ID: "catalog", Generation: 7},
				{ID: "iam_system_resources", Generation: 7},
				{ID: "workflow", Generation: 7},
			}),
		reflect.ValueOf(
			[]platformtx.CommandAuthorityDomain{
				{Name: "analytics", Generation: 7},
				{Name: "billing", Generation: 7},
				{Name: "forms", Generation: 7},
				{Name: "catalog", Generation: 7},
				{Name: "iam", Generation: 7},
				{Name: "system_resource_layout", Generation: 7},
				{Name: "workflow", Generation: 7},
			}),
		reflect.ValueOf(
			[]string{"analytics", "catalog", "forms", "iam", "workflow"},
		),
	})
	if len(output) != 2 || !output[1].IsNil() {
		t.Fatalf("admit Catalog demo scope: %v", output[1].Interface())
	}
	request, ok := output[0].Interface().(platformtx.ScopeRequest)
	if !ok {
		t.Fatal("Catalog demo scope request ABI differs")
	}
	return request
}

func catalogDemoUUIDValue(
	t *testing.T, target reflect.Type, value string,
) reflect.Value {
	t.Helper()
	raw, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	if err != nil || len(raw) != 16 ||
		target.Kind() != reflect.Array || target.Len() != 16 {
		t.Fatalf("Catalog demo UUID %q differs: %v", value, err)
	}
	result := reflect.New(target).Elem()
	for index, octet := range raw {
		result.Index(index).SetUint(uint64(octet))
	}
	return result
}

func assertCatalogDemoPending(
	t *testing.T, label string, receipt catalogDemoProvisionReceipt,
) {
	t.Helper()
	if !catalogDemoUUIDv7(receipt.ID) || receipt.Ready ||
		receipt.State != "AVAILABLE" || receipt.RetryAt.IsZero() ||
		receipt.CompletedAt.IsZero() ||
		!receipt.RetryAt.After(receipt.CompletedAt) {
		t.Fatalf("%s pending receipt differs: %+v", label, receipt)
	}
}

func assertCatalogDemoReadyReplay(
	t *testing.T,
	label string,
	pending catalogDemoProvisionReceipt,
	ready catalogDemoProvisionReceipt,
) {
	t.Helper()
	if !ready.Ready || ready.State != "AVAILABLE" || !ready.RetryAt.IsZero() ||
		ready.CompletedAt.IsZero() || ready.ID != pending.ID ||
		ready.CompletedAt.Before(pending.CompletedAt) {
		t.Fatalf("%s ready replay differs: pending=%+v ready=%+v",
			label, pending, ready)
	}
}

func assertCatalogDemoExactReadyReplay(
	t *testing.T,
	ready catalogDemoProvisionReceipt,
	replay catalogDemoProvisionReceipt,
) {
	t.Helper()
	if ready != replay {
		t.Fatalf("Catalog exact ready replay differs: ready=%+v replay=%+v",
			ready, replay)
	}
}
