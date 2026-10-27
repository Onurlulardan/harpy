//go:build integration

package executor

import "testing"

func (gate *pg18Gate) v1OffsetOracle(
	t *testing.T, offset int,
) []string {
	t.Helper()
	rows, err := gate.admin.Query(gate.ctx, `SELECT fld_id::text
FROM tenant_a.frm_executor_fixture
ORDER BY fld_name COLLATE pg_catalog.pg_unicode_fast ASC NULLS LAST,
fld_id ASC NULLS LAST LIMIT 2 OFFSET $1`, offset)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	result := make([]string, 0, 2)
	for rows.Next() {
		var identifier string
		if err := rows.Scan(&identifier); err != nil {
			t.Fatal(err)
		}
		result = append(result, identifier)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return result
}

func pg18ResultIDs(t *testing.T, result Result) []string {
	t.Helper()
	documents := pg18Documents(t, result)
	identifiers := make([]string, len(documents))
	for index, document := range documents {
		value, ok := document["id"].(string)
		if !ok {
			t.Fatalf("document id=%T", document["id"])
		}
		identifiers[index] = value
	}
	return identifiers
}
