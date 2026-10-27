//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

type billingLicenseFixture struct {
	licenseID, organizationID, environmentID uuid.UUID
	issuer, keyID, channel                   string
}

type billingLicenseClaim struct {
	operationID, leaseToken uuid.UUID
	nonce                   []byte
	headFound               bool
	headSequence            int64
	headDigest              string
	databaseNow, leaseUntil time.Time
}

type billingLicenseSnapshot struct {
	sequence                          int64
	status                            string
	issuedAt, notBefore, refreshAfter time.Time
	grantUntil, graceUntil, expiresAt time.Time
	canonical, signed                 []byte
	digest                            string
}

func newBillingLicenseFixture() billingLicenseFixture {
	return billingLicenseFixture{
		licenseID: uuid.New(), organizationID: uuid.New(),
		environmentID: uuid.MustParse(billingEnvironmentID),
		issuer:        "harpy", keyID: "root-v1", channel: "stable",
	}
}

func newBillingLicenseSnapshot(sequence int64, status, value string) billingLicenseSnapshot {
	now := time.Now().UTC().Truncate(time.Microsecond)
	canonical := []byte(fmt.Sprintf(`{"sequence":%d,"value":%q}`, sequence, value))
	digest := sha256.Sum256(canonical)
	return billingLicenseSnapshot{
		sequence: sequence, status: status, issuedAt: now.Add(-time.Hour),
		notBefore: now.Add(-time.Hour), refreshAfter: now.Add(time.Hour),
		grantUntil: now.Add(2 * time.Hour), graceUntil: now.Add(3 * time.Hour),
		expiresAt: now.Add(4 * time.Hour), canonical: canonical,
		signed: []byte(`{"signature":"fixture"}`), digest: hex.EncodeToString(digest[:]),
	}
}

func claimBillingLicense(
	ctx context.Context, connection *sql.Conn, fixture billingLicenseFixture,
	operationID uuid.UUID, leaseSeconds int,
) (billingLicenseClaim, error) {
	claim := billingLicenseClaim{}
	var licenseID, organizationID, environmentID uuid.UUID
	var delivery string
	err := connection.QueryRowContext(ctx, `SELECT * FROM
platform.billing_license_refresh_claim_v1($1,$2,$3,$4,$5,$6,$7)`,
		operationID, fixture.licenseID, fixture.organizationID,
		fixture.environmentID, "connected",
		`[{"component":"core-worker","version":"v2.0.0"}]`, leaseSeconds,
	).Scan(&claim.operationID, &claim.leaseToken, &licenseID, &organizationID,
		&environmentID, &delivery, &claim.nonce, &claim.headFound,
		&claim.headSequence, &claim.headDigest, &claim.databaseNow, &claim.leaseUntil)
	return claim, err
}

func acceptBillingLicense(
	ctx context.Context, connection *sql.Conn, fixture billingLicenseFixture,
	claim billingLicenseClaim, snapshot billingLicenseSnapshot,
) (string, error) {
	var outcome string
	err := connection.QueryRowContext(ctx, `SELECT
platform.billing_license_snapshot_accept_v1(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,
$19::jsonb,$20::jsonb,$21,$22,$23::jsonb)`,
		claim.operationID, claim.leaseToken, fixture.licenseID,
		fixture.organizationID, fixture.environmentID, "connected",
		fixture.issuer, fixture.keyID, snapshot.digest, snapshot.status,
		snapshot.sequence, snapshot.issuedAt, snapshot.notBefore,
		snapshot.refreshAfter, snapshot.grantUntil, snapshot.graceUntil,
		snapshot.expiresAt, fixture.channel, `{}`, `{}`, snapshot.canonical,
		snapshot.signed, billingLicenseProjection(),
	).Scan(&outcome)
	return outcome, err
}

func failBillingLicense(
	ctx context.Context, connection *sql.Conn,
	claim billingLicenseClaim, kind string,
) (string, error) {
	var state string
	err := connection.QueryRowContext(ctx, `SELECT
platform.billing_license_refresh_failure_v1($1,$2,$3)`,
		claim.operationID, claim.leaseToken, kind).Scan(&state)
	return state, err
}

func billingLicenseProjection() string {
	type grant struct {
		Key       string `json:"entitlement_key"`
		Type      string `json:"value_type"`
		Boolean   *bool  `json:"boolean_value"`
		Integer   *int64 `json:"integer_value"`
		Unlimited bool   `json:"unlimited"`
	}
	truth, ten := true, int64(10)
	booleans := []string{"analytics.dashboard_builder.enabled", "api.access",
		"api.builder.enabled", "workflow.builder.enabled", "workflow.enabled"}
	limits := []string{"analytics.active_dashboards", "api.active_endpoints", "documents.storage_bytes",
		"iam.active_users", "workflow.active_definitions"}
	grants := make([]grant, 0, 10)
	for _, key := range booleans {
		grants = append(grants, grant{Key: key, Type: "boolean", Boolean: &truth})
	}
	for _, key := range limits {
		if key == "iam.active_users" {
			grants = append(grants, grant{Key: key, Type: "integer", Unlimited: true})
			continue
		}
		grants = append(grants, grant{Key: key, Type: "integer", Integer: &ten})
	}
	document, _ := json.Marshal(grants)
	return string(document)
}

func countBillingLicenseRows(
	t *testing.T, ctx context.Context, connection *sql.Conn, table string,
) int {
	t.Helper()
	var count int
	query := "SELECT count(*) FROM platform." + table
	if err := connection.QueryRowContext(ctx, query).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
