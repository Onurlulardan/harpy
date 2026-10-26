//go:build integration

package postgres

import (
	"context"
	"errors"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type iamGuardRaceResult struct {
	count    int
	revision int64
	err      error
}

func runIAMGuardRace(call func() iamGuardRaceResult) []iamGuardRaceResult {
	start := make(chan struct{})
	results := make(chan iamGuardRaceResult, 2)
	var wait sync.WaitGroup
	for range 2 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			<-start
			results <- call()
		}()
	}
	close(start)
	wait.Wait()
	close(results)
	collected := make([]iamGuardRaceResult, 0, 2)
	for result := range results {
		collected = append(collected, result)
	}
	sort.Slice(collected, func(i, j int) bool { return collected[i].count < collected[j].count })
	return collected
}

func assertIAMThrottleConcurrency(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	tenantResults := runIAMGuardRace(func() iamGuardRaceResult {
		count, _, revision, err := recordIAMTenantAttempt(ctx, database.runtime,
			iamGuardTenantA, "tenant-race", iamBytes(71, 32), 2,
			iamGuardPolicyVersion, iamGuardAuthWindow,
			iamGuardIdentifierThreshold, iamGuardAuthBlock)
		return iamGuardRaceResult{count: count, revision: revision, err: err}
	})
	assertIAMAttemptRace(t, "tenant", tenantResults)
	environmentResults := runIAMGuardRace(func() iamGuardRaceResult {
		count, _, revision, err := recordIAMEnvironmentAttempt(ctx, database.runtime,
			"environment-race", iamBytes(72, 32), iamGuardPolicyVersion,
			iamGuardAuthWindow, iamGuardIdentifierThreshold, iamGuardAuthBlock)
		return iamGuardRaceResult{count: count, revision: revision, err: err}
	})
	assertIAMAttemptRace(t, "environment", environmentResults)
}

func assertIAMAttemptRace(
	t *testing.T, label string, results []iamGuardRaceResult,
) {
	t.Helper()
	if len(results) != 2 || results[0].err != nil || results[1].err != nil ||
		results[0].count != 1 || results[0].revision != 1 ||
		results[1].count != 2 || results[1].revision != 2 {
		t.Fatalf("%s attempt race=%+v", label, results)
	}
}

func assertIAMLocatorDoubleAdvance(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	now := time.Now().UTC()
	selector := iamBytes(73, 16)
	err := createIAMGuardLocator(ctx, database.runtime, selector, iamGuardTenantA, 2,
		"active", now.Add(-time.Minute), now, now.Add(48*time.Hour),
		now.Add(73*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	results := runIAMGuardRace(func() iamGuardRaceResult {
		revision, err := advanceIAMGuardLocator(ctx, database.runtime, selector,
			iamGuardTenantA, 2, 1, 2, "consumed", now.Add(time.Minute),
			now.Add(74*time.Hour))
		return iamGuardRaceResult{count: int(revision), revision: revision, err: err}
	})
	success, stale := 0, 0
	for _, result := range results {
		if result.err == nil && result.revision == 2 {
			success++
		} else if iamGuardSQLState(result.err) == "P5711" {
			stale++
		}
	}
	if success != 1 || stale != 1 {
		t.Fatalf("locator double advance success/stale=%d/%d", success, stale)
	}
}

func iamGuardSQLState(err error) string {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		return postgresError.Code
	}
	return ""
}
