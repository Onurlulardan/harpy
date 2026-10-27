//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
)

func assertBillingV1OfferingLifecycle(
	t *testing.T, ctx context.Context, connection *sql.Conn,
) {
	t.Helper()
	tests := []struct {
		key       string
		lifecycle string
		available bool
		ended     bool
	}{
		{key: "base-plan", lifecycle: "active", available: true},
		{key: "hidden-plan", lifecycle: "active", available: false},
		{key: "deleted-plan", lifecycle: "retired", available: false, ended: true},
	}
	for _, test := range tests {
		var lifecycle string
		var available, ended bool
		err := connection.QueryRowContext(ctx, `
SELECT root.lifecycle, version_row.available,
       version_row.effective_end IS NOT NULL
FROM platform.billing_offerings AS root
JOIN platform.billing_offering_versions AS version_row
  ON version_row.authority_id = root.authority_id
 AND version_row.offering_id = root.id
WHERE root.offering_key = $1`, test.key,
		).Scan(&lifecycle, &available, &ended)
		if err != nil {
			t.Fatal(err)
		}
		if lifecycle != test.lifecycle || available != test.available || ended != test.ended {
			t.Fatalf("%s lifecycle/available/ended = %s/%t/%t",
				test.key, lifecycle, available, ended)
		}
	}
}
