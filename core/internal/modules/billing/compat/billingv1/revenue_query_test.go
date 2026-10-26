package billingv1

import (
	"testing"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestRevenueReportQueryNeverSumsAcrossCurrenciesOrRegistries(t *testing.T) {
	handler := &Handler{query: billingapi.NewQueryDecoder()}
	query, err := handler.revenueReportQuery(billingapi.Access{})
	if err != nil {
		t.Fatal(err)
	}
	aggregate, found := query.Request.Aggregate()
	if !found || aggregate.Limit() != 500 {
		t.Fatalf("aggregate found=%t limit=%d", found, aggregate.Limit())
	}
	groups := aggregate.GroupBy()
	if len(groups) != 2 || groups[0] != "currency" ||
		groups[1] != "currency_registry_version" {
		t.Fatalf("revenue groups=%v", groups)
	}
	measures := aggregate.Measures()
	if len(measures) != 2 || measures[0].Name() != "monthly_minor_units" ||
		measures[1].Name() != "annual_minor_units" {
		t.Fatalf("revenue measures=%v", measures)
	}
	sorts := aggregate.Sort()
	if len(sorts) != 2 || sorts[0].Name() != "currency" ||
		sorts[1].Name() != "currency_registry_version" {
		t.Fatalf("revenue sorts=%v", sorts)
	}
}
