package requestmeta

import (
	"context"
	"testing"
)

func TestCorrelationIDContextRoundTrip(t *testing.T) {
	if _, found := CorrelationIDFromContext(context.Background()); found {
		t.Fatal("unexpected correlation ID in empty context")
	}
	ctx := WithCorrelationID(context.Background(), "request-1")
	resolved, found := CorrelationIDFromContext(ctx)
	if !found || resolved != "request-1" {
		t.Fatalf("unexpected context correlation ID: %q %v", resolved, found)
	}
}
