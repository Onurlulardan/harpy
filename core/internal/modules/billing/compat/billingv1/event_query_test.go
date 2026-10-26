package billingv1

import (
	"errors"
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestEventReportQueryPreservesV1LimitOrderAndTenantFilter(t *testing.T) {
	handler := &Handler{query: billingapi.NewQueryDecoder()}
	tenantID := "11111111-1111-4111-8111-111111111111"
	query, err := handler.eventReportQuery(billingapi.Access{}, tenantID)
	if err != nil || query.Request.Page().Size() != 100 {
		t.Fatalf("event query page=%d error=%v", query.Request.Page().Size(), err)
	}
	sorts := query.Request.Sort()
	if len(sorts) != 1 || sorts[0].Field() != "created_at" ||
		sorts[0].Direction().String() != "desc" {
		t.Fatalf("event sorts=%v", sorts)
	}
	where, found := query.Request.Where()
	if !found {
		t.Fatal("event tenant filter is missing")
	}
	children := where.Children()
	if len(children) != 1 {
		t.Fatalf("event tenant filter children=%d", len(children))
	}
	predicate, predicateFound := children[0].Predicate()
	if !predicateFound || predicate.Field() != "tenant_id" {
		t.Fatalf("event tenant filter=%v found=%t/%t", where, found, predicateFound)
	}
	value, text, valueErr := predicate.Values()[0].TextValue()
	if valueErr != nil || !text || value != tenantID {
		t.Fatalf("event tenant literal=%q text=%t error=%v", value, text, valueErr)
	}
}

func TestEventReportQueryRejectsNonemptyInvalidTenant(t *testing.T) {
	handler := &Handler{query: billingapi.NewQueryDecoder()}
	_, err := handler.eventReportQuery(billingapi.Access{}, "not-a-tenant")
	if !errors.Is(err, ErrRequestInvalid) {
		t.Fatalf("invalid tenant error=%v", err)
	}
}
