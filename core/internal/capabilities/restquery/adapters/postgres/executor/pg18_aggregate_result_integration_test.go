//go:build integration

package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type pg18AggregateDocument struct {
	Key      map[string]json.RawMessage `json:"key"`
	Measures map[string]json.RawMessage `json:"measures"`
}

func (gate *pg18Gate) runAggregate(
	t *testing.T, executor Executor, plan resolved.AuthorizedPlan,
) (AggregateResult, error) {
	t.Helper()
	var result AggregateResult
	err := gate.coordinator.InTenant(
		gate.ctx, pg18ScopeRequest(t, pg18TenantA, "forms"),
		func(ctx context.Context, scope transaction.TenantScope) error {
			next, executeErr := executor.ExecuteAggregate(ctx, scope, plan)
			if executeErr == nil {
				result = next
			}
			return executeErr
		},
	)
	return result, err
}

func pg18AggregateDocuments(
	t *testing.T, result AggregateResult,
) []pg18AggregateDocument {
	t.Helper()
	groups, err := result.Groups()
	if err != nil {
		t.Fatal(err)
	}
	documents := make([]pg18AggregateDocument, len(groups))
	for index := range groups {
		if err := json.Unmarshal(groups[index], &documents[index]); err != nil {
			t.Fatal(err)
		}
	}
	return documents
}

func pg18JSONString(document json.RawMessage) (string, bool) {
	if bytes.Equal(document, []byte("null")) || len(document) == 0 {
		return "", false
	}
	var value string
	return value, json.Unmarshal(document, &value) == nil
}
