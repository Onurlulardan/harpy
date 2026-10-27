//go:build integration

package integrationtest

import (
	"testing"

	"github.com/google/uuid"
)

func TestBillingInstallationOfferingRuntimeIdentityAgainstPostgreSQL184(
	t *testing.T,
) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := billingTestContext(t)
			defer cancel()
			database := prepareBillingOrigin(t, ctx, origin)
			connection := billingRuntimeConnection(t, ctx, database)
			if _, err := connection.ExecContext(
				ctx, "SET ROLE harpy_runtime",
			); err != nil {
				t.Fatal(err)
			}
			_, err := connection.ExecContext(ctx, `SELECT * FROM
platform.billing_installation_offering_read_v1($1,'onpremise')`, uuid.New())
			if billingPGCode(err) != "55000" {
				t.Fatalf("runtime offering read code=%q err=%v",
					billingPGCode(err), err)
			}
			_, err = database.ExecContext(ctx, `SELECT * FROM
platform.billing_installation_offering_read_v1($1,'onpremise')`, uuid.New())
			if billingPGCode(err) != "42501" {
				t.Fatalf("owner offering read code=%q err=%v",
					billingPGCode(err), err)
			}
		})
	}
}
