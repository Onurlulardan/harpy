//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

const billingLicenseManifest = "3a0e120dd180d30f0e9c4ccb20c5338554c5f9e6db8349bff5335d813a421e99"

type billingOfferingReceipt struct {
	operationID, versionID uuid.UUID
	revision               int64
	digest                 string
	replay                 bool
}

func assertBillingLicenseOffering(
	t *testing.T, ctx context.Context, connection *sql.Conn,
	state billingLicenseAcceptedState,
) {
	t.Helper()
	effectiveAt := time.Now().UTC().Truncate(time.Microsecond)
	versionID, operationID := billingV7(t), billingV7(t)
	document := billingLicenseOfferingDocument(state, versionID, effectiveAt)
	digest := sha256.Sum256(document)
	digestText := hex.EncodeToString(digest[:])
	missing := state
	missing.fixture.licenseID = uuid.New()
	missingDocument := billingLicenseOfferingDocument(missing, billingV7(t), effectiveAt)
	missingDigest := sha256.Sum256(missingDocument)
	_, err := publishBillingLicenseOffering(
		ctx, connection, missing, billingV7(t), billingV7(t), effectiveAt,
		missingDocument, hex.EncodeToString(missingDigest[:]),
	)
	if billingPGCode(err) != "55000" {
		t.Fatalf("unverified offering code=%q err=%v", billingPGCode(err), err)
	}
	receipt, err := publishBillingLicenseOffering(
		ctx, connection, state, operationID, versionID, effectiveAt, document, digestText,
	)
	if err != nil || receipt.replay || receipt.versionID != versionID {
		t.Fatalf("license offering receipt=%+v err=%v", receipt, err)
	}
	replay, err := publishBillingLicenseOffering(
		ctx, connection, state, operationID, versionID, effectiveAt, document, digestText,
	)
	if err != nil || !replay.replay {
		t.Fatalf("license offering replay=%+v err=%v", replay, err)
	}
	assertBillingOfferingCounts(t, ctx, connection, versionID)
}

func publishBillingLicenseOffering(
	ctx context.Context, connection *sql.Conn, state billingLicenseAcceptedState,
	operationID, versionID uuid.UUID, effectiveAt time.Time,
	document []byte, digest string,
) (billingOfferingReceipt, error) {
	receipt := billingOfferingReceipt{}
	fixture, snapshot := state.fixture, state.snapshot
	err := connection.QueryRowContext(ctx, `SELECT * FROM
platform.billing_publish_license_derived_offering_v1(
$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		operationID, fixture.environmentID, fixture.licenseID,
		fixture.organizationID, "onpremise", "connected", fixture.issuer,
		snapshot.sequence, snapshot.digest, versionID, int64(1), digest,
		effectiveAt, billingLicenseManifest, document,
	).Scan(&receipt.operationID, &receipt.versionID, &receipt.revision,
		&receipt.digest, &receipt.replay)
	return receipt, err
}

func billingLicenseOfferingDocument(
	state billingLicenseAcceptedState, versionID uuid.UUID, effectiveAt time.Time,
) []byte {
	type booleanGrant struct {
		Key   string `json:"key"`
		Value bool   `json:"value"`
	}
	type integerGrant struct {
		Key       string `json:"key"`
		Value     int64  `json:"value"`
		Unlimited bool   `json:"unlimited"`
	}
	booleans := []booleanGrant{}
	for _, key := range []string{"analytics.dashboard_builder.enabled", "api.access",
		"api.builder.enabled", "forms.builder.enabled", "workflow.builder.enabled", "workflow.enabled"} {
		booleans = append(booleans, booleanGrant{Key: key, Value: true})
	}
	limits := []integerGrant{}
	for _, key := range []string{"analytics.active_dashboards", "api.active_endpoints",
		"documents.storage_bytes", "forms.active_forms", "forms.submissions",
		"iam.active_users", "workflow.active_definitions"} {
		grant := integerGrant{Key: key, Value: 10}
		if key == "iam.active_users" {
			grant.Value, grant.Unlimited = 0, true
		}
		limits = append(limits, grant)
	}
	document := map[string]any{
		"schema":              "harpy-license-derived-offering.v1",
		"offering_version_id": versionID.String(), "revision": int64(1),
		"effective_at":             effectiveAt.Format(time.RFC3339Nano),
		"license_id":               state.fixture.licenseID.String(),
		"licensee_organization_id": state.fixture.organizationID.String(),
		"environment_id":           state.fixture.environmentID.String(),
		"issuer":                   state.fixture.issuer, "snapshot_sequence": state.snapshot.sequence,
		"snapshot_sha256": state.snapshot.digest, "release_channel": state.fixture.channel,
		"booleans": booleans, "limits": limits,
	}
	encoded, _ := json.Marshal(document)
	return encoded
}

func assertBillingOfferingCounts(
	t *testing.T, ctx context.Context, connection *sql.Conn, versionID uuid.UUID,
) {
	t.Helper()
	checks := []struct {
		query string
		want  int
	}{
		{"SELECT count(*) FROM platform.billing_license_offering_publications WHERE $1::uuid IS NOT NULL", 1},
		{`SELECT count(*) FROM platform.billing_offerings
WHERE id=(SELECT offering_id FROM platform.billing_offering_versions WHERE id=$1)`, 1},
		{"SELECT count(*) FROM platform.billing_offering_versions WHERE id=$1", 1},
		{"SELECT count(*) FROM platform.billing_price_versions WHERE offering_version_id=$1", 1},
		{"SELECT count(*) FROM platform.billing_offering_entitlements WHERE offering_version_id=$1", 14},
		{"SELECT count(*) FROM platform.billing_tenant_base_offerings WHERE offering_version_id=$1", 0},
		{"SELECT count(*) FROM platform.billing_subscriptions WHERE offering_version_id=$1", 0},
		{"SELECT count(*) FROM platform.billing_trial_claims WHERE offering_version_id=$1", 0},
	}
	for _, check := range checks {
		var got int
		err := connection.QueryRowContext(ctx, check.query, versionID).Scan(&got)
		if err != nil || got != check.want {
			t.Fatalf("offering count query=%q got=%d want=%d err=%v",
				strings.TrimSpace(check.query), got, check.want, err)
		}
	}
}
