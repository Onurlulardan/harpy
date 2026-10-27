//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

type meteredAdmissionSQL struct {
	lockSnapshot string
	decision     string
	entry        string
}

func loadMeteredAdmissionSQL(t *testing.T) meteredAdmissionSQL {
	t.Helper()
	return meteredAdmissionSQL{
		lockSnapshot: loadBillingAdapterSQL(t, "usage", "lock_snapshot.sql"),
		decision:     loadBillingAdapterSQL(t, "usage", "insert_decision.sql"),
		entry:        loadBillingAdapterSQL(t, "usage", "insert_gauge_entry.sql"),
	}
}

func writeMeteredAdmission(
	ctx context.Context,
	transaction *sql.Tx,
	fixture meteredResourceFixture,
	statements meteredAdmissionSQL,
	operationID, ownerFactID uuid.UUID,
	idempotencyKey string,
) (bool, error) {
	var receiptFound, featureEnabled, priorAllowed, priorUnlimited bool
	var priorDelta, priorCurrent, priorLimit, projectionRevision int64
	var projectionFound bool
	var nullableOne, nullableTwo, nullableThree, nullableFour any
	if err := transaction.QueryRowContext(ctx, `
SELECT * FROM platform.billing_metered_resource_inspect_v1(
 $1,$2,$3,'api_endpoint','create',$4,$5
)`, fixture.authorityID, fixture.tenantID, fixture.actorID,
		idempotencyKey, ownerFactID,
	).Scan(&receiptFound, &nullableOne, &nullableTwo, &nullableThree,
		&featureEnabled, &priorAllowed, &priorDelta, &priorCurrent, &priorLimit,
		&priorUnlimited, &nullableFour, &projectionFound, &nullableOne,
		&projectionRevision, &nullableTwo); err != nil {
		return false, fmt.Errorf("inspect resource: %w", err)
	}
	if receiptFound || projectionFound {
		return false, fmt.Errorf("unexpected metered resource replay")
	}
	var enabled bool
	if err := transaction.QueryRowContext(ctx, `
SELECT platform.billing_metered_resource_feature_v1(
 $1,$2,'api.builder.enabled'
)`, fixture.authorityID, fixture.tenantID).Scan(&enabled); err != nil || !enabled {
		return false, fmt.Errorf("resolve builder feature enabled=%v: %w", enabled, err)
	}

	var key, valueType, meterKind, resetPolicy, enforcement string
	var unlimitedAllowed, unlimited bool
	var limit sql.NullInt64
	var periodID uuid.UUID
	var periodStart, periodEnd time.Time
	var committed, reserved, current, revision int64
	if err := transaction.QueryRowContext(ctx, statements.lockSnapshot,
		fixture.authorityID, fixture.tenantID, "api.active_endpoints",
		fixture.periodID, "gauge",
	).Scan(&key, &valueType, &meterKind, &resetPolicy, &enforcement,
		&unlimitedAllowed, &limit, &unlimited, &periodID, &periodStart, &periodEnd,
		&committed, &reserved, &current, &revision); err != nil {
		return false, fmt.Errorf("lock resource gauge: %w", err)
	}
	allowed := unlimited || limit.Valid && current < limit.Int64
	limitArgument := any(nil)
	publicLimit := int64(0)
	if limit.Valid {
		limitArgument, publicLimit = limit.Int64, limit.Int64
	}
	decisionID := uuid.Must(uuid.NewV7())
	var storedDecision uuid.UUID
	if err := transaction.QueryRowContext(ctx, statements.decision,
		fixture.authorityID, fixture.tenantID, fixture.actorID, decisionID,
		"api.active_endpoints", fixture.periodID, operationID, int64(1), current,
		limitArgument, unlimited, allowed, false, fixture.now,
	).Scan(&storedDecision); err != nil {
		return false, fmt.Errorf("insert resource gauge decision: %w", err)
	}
	var entryArgument any
	if allowed {
		entryID := uuid.Must(uuid.NewV7())
		var storedEntry uuid.UUID
		if err := transaction.QueryRowContext(ctx, statements.entry,
			fixture.authorityID, fixture.tenantID, entryID, decisionID,
			"api.active_endpoints", fixture.periodID, int64(1), current, current+1,
			fixture.now, operationID,
		).Scan(&storedEntry); err != nil {
			return false, fmt.Errorf("insert resource gauge entry: %w", err)
		}
		entryArgument = storedEntry
	}
	var recorded bool
	if err := transaction.QueryRowContext(ctx, `
SELECT platform.billing_metered_resource_record_v1(
 $1,$2,$3,'api_endpoint','create',$4,$5,$6,$7,true,$8,$9,$10,$11,$12,$13,$14,0,$15
)`, fixture.authorityID, fixture.tenantID, fixture.actorID, operationID,
		ownerFactID, idempotencyKey, billingDigest(idempotencyKey), allowed,
		map[bool]int64{true: 1, false: 0}[allowed], current, publicLimit, unlimited,
		storedDecision, entryArgument, fixture.now,
	).Scan(&recorded); err != nil {
		return false, fmt.Errorf("record resource admission: %w", err)
	}
	if !recorded {
		return false, fmt.Errorf("resource admission was not recorded")
	}
	return allowed, nil
}
