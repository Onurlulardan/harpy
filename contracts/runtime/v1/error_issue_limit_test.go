package runtimev1

import (
	"strings"
	"testing"
)

func TestNewErrorEnvelopeEnforcesTwentyIssueLimit(t *testing.T) {
	t.Parallel()
	issues := make([]Issue, maximumErrorIssues)
	for index := range issues {
		issues[index] = Issue{Path: "/field", Code: "FIELD_UNKNOWN"}
	}
	if _, err := NewErrorEnvelope(
		"QUERY_INVALID", "Invalid query", "request-1", issues,
	); err != nil {
		t.Fatalf("twenty issues rejected: %v", err)
	}
	issues = append(issues, Issue{Path: "/field", Code: "FIELD_UNKNOWN"})
	if _, err := NewErrorEnvelope(
		"QUERY_INVALID", "Invalid query", "request-1", issues,
	); err == nil || !strings.Contains(err.Error(), "zero to 20") {
		t.Fatalf("twenty-one issues were not rejected by the hard limit: %v", err)
	}
}

func TestErrorEnvelopeSchemaEnforcesTwentyIssueLimit(t *testing.T) {
	t.Parallel()
	issues := make([]any, maximumErrorIssues+1)
	for index := range issues {
		issues[index] = map[string]any{"path": "/field", "code": "FIELD_UNKNOWN"}
	}
	wire := map[string]any{
		"success": false,
		"error": map[string]any{
			"code": "QUERY_INVALID", "message": "Invalid query",
			"issues": issues, "correlation_id": "request-1",
		},
	}
	if err := compileSchema(t, "error-envelope.schema.json").Validate(wire); err == nil {
		t.Fatal("schema accepted twenty-one issues")
	}
}
