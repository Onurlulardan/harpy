package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type IAMThrottleProbe struct {
	ScopeKind, Purpose, Dimension, KeyID string
	TenantID                             uuid.UUID
	BucketHMAC                           [32]byte
	GroupGeneration, IAMGeneration       int64
}

func (pool *Pool) IAMThrottleBlocked(
	ctx context.Context, scopeKind string, tenantID uuid.UUID,
	purpose, dimension, keyID string, bucketHMAC [32]byte,
	groupGeneration, iamGeneration int64,
) (bool, error) {
	return pool.executeIAMThrottle(ctx, IAMThrottleProbe{
		ScopeKind: scopeKind, TenantID: tenantID, Purpose: purpose,
		Dimension: dimension, KeyID: keyID, BucketHMAC: bucketHMAC,
		GroupGeneration: groupGeneration, IAMGeneration: iamGeneration,
	}, false)
}

func (pool *Pool) IAMThrottleRecordFailure(
	ctx context.Context, scopeKind string, tenantID uuid.UUID,
	purpose, dimension, keyID string, bucketHMAC [32]byte,
	groupGeneration, iamGeneration int64,
) (bool, error) {
	return pool.executeIAMThrottle(ctx, IAMThrottleProbe{
		ScopeKind: scopeKind, TenantID: tenantID, Purpose: purpose,
		Dimension: dimension, KeyID: keyID, BucketHMAC: bucketHMAC,
		GroupGeneration: groupGeneration, IAMGeneration: iamGeneration,
	}, true)
}

func (pool *Pool) executeIAMThrottle(
	ctx context.Context, probe IAMThrottleProbe, record bool,
) (bool, error) {
	if pool == nil || ctx == nil || !probe.valid() {
		return false, fmt.Errorf("IAM throttle probe is invalid")
	}
	mode := TransactionSnapshotQuery
	query := iamThrottleBlockedSQL
	if record {
		mode, query = TransactionCommand, iamThrottleRecordFailureSQL
	}
	database, err := pool.BeginRuntimeTransaction(ctx, mode)
	if err != nil {
		return false, err
	}
	var tenantID any
	var groupGeneration, iamGeneration any
	if probe.ScopeKind == "tenant" {
		tenantID = probe.TenantID
		groupGeneration, iamGeneration = probe.GroupGeneration, probe.IAMGeneration
	}
	var blocked bool
	err = database.transaction.QueryRow(
		ctx, query, probe.ScopeKind, tenantID, probe.Purpose, probe.Dimension,
		probe.KeyID, probe.BucketHMAC[:], groupGeneration, iamGeneration,
	).Scan(&blocked)
	if err != nil {
		_ = database.Rollback()
		return false, err
	}
	if err := database.Commit(ctx); err != nil {
		return false, err
	}
	return blocked, nil
}

func (probe IAMThrottleProbe) valid() bool {
	if probe.Purpose != "login" && probe.Purpose != "refresh" ||
		probe.KeyID == "" || probe.BucketHMAC == [32]byte{} {
		return false
	}
	if probe.ScopeKind == "environment" {
		return probe.TenantID == uuid.Nil && probe.GroupGeneration == 0 &&
			probe.IAMGeneration == 0
	}
	return probe.ScopeKind == "tenant" && probe.TenantID != uuid.Nil &&
		probe.GroupGeneration > 0 && probe.IAMGeneration > 0
}
