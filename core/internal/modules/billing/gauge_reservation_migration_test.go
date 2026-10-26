package billing

import (
	"strings"
	"testing"
)

func TestGaugeReservationConstraintIsForwardOnlyAndValidated(t *testing.T) {
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	const name = "20260718200050_allow_billing_gauge_reservations_v1"
	found := false
	for _, document := range documents {
		if !strings.Contains(document.Path(), "/fresh/"+name+".fresh.sql") {
			continue
		}
		found = true
		text := string(document.Content())
		for _, required := range []string{
			"DROP CONSTRAINT billing_usage_totals_value_check",
			"meter_kind = 'gauge' AND committed_total = 0",
			") NOT VALID", "VALIDATE CONSTRAINT billing_usage_totals_value_check",
		} {
			if !strings.Contains(text, required) {
				t.Fatalf("gauge reservation migration misses %q", required)
			}
		}
	}
	if !found {
		t.Fatal("gauge reservation migration is missing")
	}
}
