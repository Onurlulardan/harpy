package postgres

import (
	"testing"
	"time"
)

func TestTraceSpanLookupScansOwnedSpan(t *testing.T) {
	now := time.Unix(1, 0).UTC()
	database := &operationDatabase{row: func(destinations ...any) error {
		*destinations[0].(*string) = "row"
		*destinations[1].(*string) = "environment"
		*destinations[2].(*string) = "request-1"
		*destinations[3].(*string) = "0123456789abcdef"
		*destinations[4].(**string) = nil
		*destinations[5].(*string) = "core-api"
		*destinations[6].(*string) = "read"
		*destinations[7].(*string) = "server"
		*destinations[8].(*string) = "GET"
		*destinations[9].(*string) = "/api/example"
		*destinations[10].(*int) = 200
		*destinations[11].(*time.Time) = now
		*destinations[12].(*time.Time) = now.Add(time.Second)
		*destinations[13].(*float64) = 1000
		*destinations[14].(*string) = "info"
		*destinations[15].(**string) = nil
		*destinations[16].(*time.Time) = now.Add(time.Second)
		return nil
	}}
	result, err := operationStore(database).TraceSpanLookup(
		t.Context(), "environment", "0123456789abcdef")
	if err != nil || result.SpanID != "0123456789abcdef" ||
		result.EnvironmentID != "environment" {
		t.Fatalf("TraceSpanLookup() = %#v, %v", result, err)
	}
}
