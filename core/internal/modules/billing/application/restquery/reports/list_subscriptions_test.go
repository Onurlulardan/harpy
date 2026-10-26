package reports

import (
	"errors"
	"testing"

	queryapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestListSubscriptionsRejectsUnboundService(t *testing.T) {
	service := &Service{}
	_, err := service.ListSubscriptions(
		t.Context(), platformtx.GlobalScope{}, queryapi.QueryRequest{},
	)
	if !errors.Is(err, ErrContract) {
		t.Fatalf("ListSubscriptions error = %v, want ErrContract", err)
	}
}

func TestSubscriptionReportPinsPresentationAndEffectiveMoney(t *testing.T) {
	spec := subscriptionSpec()
	if spec.View != "billing_subscription_report_query_v2" {
		t.Fatalf("subscription report view = %q", spec.View)
	}
	wanted := map[string]string{
		"created_at": "timestamp", "updated_at": "timestamp",
		"effective_price_minor_units": "integer",
		"effective_price_currency":    "text", "offering": "json",
	}
	for _, field := range spec.Fields {
		if kind, found := wanted[field.Name]; found && kind == field.Kind {
			delete(wanted, field.Name)
		}
	}
	if len(wanted) != 0 {
		t.Fatalf("subscription report fields missing: %v", wanted)
	}
}
