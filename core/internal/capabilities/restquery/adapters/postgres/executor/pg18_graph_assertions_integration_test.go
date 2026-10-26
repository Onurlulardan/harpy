//go:build integration

package executor

import (
	"strings"
	"testing"
)

func pg18AssertGraphDocument(
	t *testing.T, result Result, expectedChild string,
) {
	t.Helper()
	documents := pg18Documents(t, result)
	total, present, err := result.ExactCount()
	if err != nil || !present || total != 2 || len(documents) != 2 {
		t.Fatalf("records=%v count=%d present=%t error=%v",
			documents, total, present, err)
	}
	if documents[0]["label"] != "Root One" || documents[1]["label"] != "Root Two" {
		t.Fatalf("root sort changed: %v", documents)
	}
	firstIncluded := pg18GraphObject(t, documents[0]["_included"])
	children := pg18GraphList(t, firstIncluded["children"])
	if len(children) != 1 || pg18GraphObject(t, children[0])["label"] != expectedChild {
		t.Fatalf("children=%v expected=%s", children, expectedChild)
	}
	childIncluded := pg18GraphObject(
		t, pg18GraphObject(t, children[0])["_included"],
	)
	items := pg18GraphList(t, childIncluded["items"])
	if expectedChild == "Alpha child" {
		if len(items) != 1 || pg18GraphObject(t, items[0])["label"] != "Alpha nested" {
			t.Fatalf("nested items=%v", items)
		}
	} else if len(items) != 0 {
		t.Fatalf("snapshot child unexpectedly gained nested rows: %v", items)
	}
	owner := pg18GraphObject(t, firstIncluded["owner"])
	if owner["label"] != "Visible owner" {
		t.Fatalf("owner=%v", owner)
	}
	secondIncluded := pg18GraphObject(t, documents[1]["_included"])
	if len(pg18GraphList(t, secondIncluded["children"])) != 0 ||
		secondIncluded["owner"] != nil {
		t.Fatalf("missing targets were not dropped/null: %v", secondIncluded)
	}
	page, _ := result.Page()
	records, _ := page.Records()
	for _, record := range records {
		raw := string(record)
		childrenAt, ownerAt := strings.Index(raw, `"children"`), strings.Index(raw, `"owner"`)
		if childrenAt < 0 || ownerAt < 0 || childrenAt > ownerAt ||
			strings.Contains(raw, `"id"`) || strings.Contains(raw, "graph_") {
			t.Fatalf("order or hidden material leaked: %s", raw)
		}
	}
}

func pg18AssertTenantBGraph(t *testing.T, result Result) {
	t.Helper()
	documents := pg18Documents(t, result)
	total, present, err := result.ExactCount()
	if err != nil || !present || total != 1 || len(documents) != 1 ||
		documents[0]["label"] != "Tenant B Root" ||
		strings.Contains(string(pg18MustFirstRecord(t, result)), "Root One") {
		t.Fatalf("tenant B result=%v total=%d present=%t error=%v",
			documents, total, present, err)
	}
}

func pg18AssertGraphOwners(t *testing.T, harness pg18GraphHarness) {
	t.Helper()
	if harness.forms.calls.Load() != 2 || harness.crm.calls.Load() != 1 ||
		harness.iam.calls.Load() != 1 || harness.observation.opened.Load() != 7 ||
		harness.observation.closed.Load() != 7 {
		t.Fatalf("owners forms/crm/iam=%d/%d/%d rows=%d/%d",
			harness.forms.calls.Load(), harness.crm.calls.Load(), harness.iam.calls.Load(),
			harness.observation.opened.Load(), harness.observation.closed.Load())
	}
}

func pg18AssertGraphSQLOrder(t *testing.T, calls []pg18Call) {
	t.Helper()
	pg18AssertSiblingExplains(t, calls, []string{
		pg18GraphRelationTable(graphRelationA),
		pg18GraphRelationTable(graphRelationB),
	})
	pg18AssertSiblingExplains(t, calls, []string{"frm_graph_a", "frm_graph_b"})
	countCalls := 0
	for _, call := range calls {
		lower := strings.ToLower(call.statement)
		if strings.Contains(lower, "count(*)") {
			countCalls++
			if !strings.Contains(call.statement, "frm_graph_root") ||
				strings.Contains(call.statement, "rel_") {
				t.Fatalf("exact count escaped root: %s", call.statement)
			}
		}
	}
	if countCalls != 2 {
		t.Fatalf("exact count explain/data calls=%d", countCalls)
	}
}

func pg18AssertSiblingExplains(t *testing.T, calls []pg18Call, tables []string) {
	t.Helper()
	firstData := len(calls)
	for index, call := range calls {
		if call.kind == "query" && pg18StatementUsesAny(call.statement, tables) {
			firstData = index
			break
		}
	}
	if firstData == len(calls) {
		t.Fatalf("sibling data SQL missing for %v", tables)
	}
	for _, table := range tables {
		found := false
		for index := 0; index < firstData; index++ {
			call := calls[index]
			if call.kind == "row" && strings.HasPrefix(call.statement, "EXPLAIN ") &&
				strings.Contains(call.statement, table) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s was not EXPLAINed before first sibling data SQL", table)
		}
	}
}

func pg18StatementUsesAny(statement string, values []string) bool {
	for _, value := range values {
		if strings.Contains(statement, value) {
			return true
		}
	}
	return false
}

func pg18GraphObject(t *testing.T, value any) map[string]any {
	t.Helper()
	result, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("value is not an object: %#v", value)
	}
	return result
}

func pg18GraphList(t *testing.T, value any) []any {
	t.Helper()
	result, ok := value.([]any)
	if !ok {
		t.Fatalf("value is not a list: %#v", value)
	}
	return result
}

func pg18MustFirstRecord(t *testing.T, result Result) []byte {
	t.Helper()
	page, err := result.Page()
	if err != nil {
		t.Fatal(err)
	}
	records, err := page.Records()
	if err != nil || len(records) == 0 {
		t.Fatalf("records=%q error=%v", records, err)
	}
	return records[0]
}
