package runtimev1

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestNewErrorEnvelopeMarshalsCanonicalWire(t *testing.T) {
	t.Parallel()
	issue, err := NewIssue("/where/and/0/field", "FIELD_UNKNOWN")
	if err != nil {
		t.Fatal(err)
	}
	envelope, err := NewErrorEnvelope(
		"QUERY_INVALID",
		"Request could not be processed",
		"6c3016c0-2b2b-4a65-8aef-749df4eb9d8a",
		[]Issue{issue},
	)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"success":false,"error":{"code":"QUERY_INVALID",` +
		`"message":"Request could not be processed","issues":[` +
		`{"path":"/where/and/0/field","code":"FIELD_UNKNOWN"}],` +
		`"correlation_id":"6c3016c0-2b2b-4a65-8aef-749df4eb9d8a"}}`
	if string(encoded) != want {
		t.Fatalf("unexpected wire\nwant: %s\n got: %s", want, encoded)
	}
}

func TestNewErrorEnvelopeAlwaysEmitsIssuesArray(t *testing.T) {
	t.Parallel()
	envelope, err := NewErrorEnvelope("ROUTE_NOT_FOUND", "Route not found", "request-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(encoded), `"issues":[]`) {
		t.Fatalf("issues must be an array: %s", encoded)
	}
}

func TestNewErrorEnvelopeCopiesIssues(t *testing.T) {
	t.Parallel()
	issues := []Issue{{Path: "/field", Code: "FIELD_UNKNOWN"}}
	envelope, err := NewErrorEnvelope("QUERY_INVALID", "Invalid query", "request-1", issues)
	if err != nil {
		t.Fatal(err)
	}
	issues[0].Code = "CHANGED"
	if envelope.Error.Issues[0].Code != "FIELD_UNKNOWN" {
		t.Fatal("constructor retained caller-owned issue storage")
	}
}

func TestNewIssueAcceptsJSONPointers(t *testing.T) {
	t.Parallel()
	tests := []string{
		"",
		"/where/and/0/field",
		"/properties/a~1b/~0flag",
	}
	for _, path := range tests {
		path := path
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			issue, err := NewIssue(path, "FIELD_UNKNOWN")
			if err != nil {
				t.Fatal(err)
			}
			if issue.Path != path {
				t.Fatalf("unexpected path: %q", issue.Path)
			}
		})
	}
}

func TestNewIssueRejectsSchemaViolations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		path string
		code string
	}{
		{name: "long path", path: strings.Repeat("ü", 257), code: "FIELD_UNKNOWN"},
		{name: "invalid utf8", path: string([]byte{0xff}), code: "FIELD_UNKNOWN"},
		{name: "overlong pointer", path: "/" + strings.Repeat("x", 256), code: "FIELD_UNKNOWN"},
		{name: "non pointer", path: "where.and[0].field", code: "FIELD_UNKNOWN"},
		{name: "trailing tilde", path: "/field~", code: "FIELD_UNKNOWN"},
		{name: "invalid escape", path: "/field~2", code: "FIELD_UNKNOWN"},
		{name: "short code", path: "/field", code: "A"},
		{name: "lowercase code", path: "/field", code: "FIELD_unknown"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewIssue(test.path, test.code); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestNewErrorEnvelopeRejectsSchemaViolations(t *testing.T) {
	t.Parallel()
	validIssues := []Issue{{Path: "/field", Code: "FIELD_UNKNOWN"}}
	tests := []struct {
		name          string
		code          string
		message       string
		correlationID string
		issues        []Issue
	}{
		{name: "code", code: "bad", message: "message", correlationID: "request-1", issues: validIssues},
		{name: "message", code: "QUERY_INVALID", message: "", correlationID: "request-1", issues: validIssues},
		{name: "correlation", code: "QUERY_INVALID", message: "message", correlationID: "bad id", issues: validIssues},
		{name: "nil issues", code: "QUERY_INVALID", message: "message", correlationID: "request-1", issues: []Issue{{Code: "bad"}}},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := NewErrorEnvelope(test.code, test.message, test.correlationID, test.issues); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestErrorEnvelopeMarshalFailsClosed(t *testing.T) {
	t.Parallel()
	envelope, err := NewErrorEnvelope("ROUTE_NOT_FOUND", "Route not found", "request-1", nil)
	if err != nil {
		t.Fatal(err)
	}
	envelope.Success = true
	if _, err := json.Marshal(envelope); err == nil {
		t.Fatal("invalid envelope crossed the JSON boundary")
	}
}
