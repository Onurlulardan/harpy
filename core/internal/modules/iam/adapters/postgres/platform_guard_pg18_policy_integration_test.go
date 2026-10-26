//go:build integration

package postgres

import (
	"context"
	"testing"
)

const iamThrottlePolicySQL = `SELECT platform.iam_auth_throttle_policy_v1(
$1,$2,$3,$4,$5,$6,$7)`

func assertIAMThrottlePolicyMatrix(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
) {
	t.Helper()
	authThreshold := map[string]int{
		"identifier": 5, "selector": 5, "identifier_ip": 10,
		"tenant_slug": 30, "ip": 100,
	}
	tenantAuth := map[string][]string{
		"login":   {"identifier", "ip", "identifier_ip"},
		"refresh": {"selector", "ip"},
		"reauth":  {"identifier", "ip", "identifier_ip"},
	}
	environmentAuth := map[string][]string{
		"login":   {"tenant_slug", "identifier", "ip", "identifier_ip"},
		"refresh": {"selector", "ip"},
	}
	for purpose, dimensions := range tenantAuth {
		for _, dimension := range dimensions {
			assertIAMThrottlePolicy(t, ctx, database, "tenant", purpose,
				dimension, 300, authThreshold[dimension], 900, true)
		}
	}
	for purpose, dimensions := range environmentAuth {
		for _, dimension := range dimensions {
			assertIAMThrottlePolicy(t, ctx, database, "environment", purpose,
				dimension, 300, authThreshold[dimension], 900, true)
		}
	}
	actionThreshold := map[string]int{
		"identifier": 3, "identifier_ip": 10, "tenant_slug": 30, "ip": 100,
	}
	for _, purpose := range []string{
		"password_reset", "email_verification", "portal_setup", "invitation",
	} {
		for _, dimension := range []string{"identifier", "ip", "identifier_ip"} {
			assertIAMThrottlePolicy(t, ctx, database, "tenant", purpose,
				dimension, 900, actionThreshold[dimension], 3600, true)
		}
		for _, dimension := range []string{
			"tenant_slug", "identifier", "ip", "identifier_ip",
		} {
			assertIAMThrottlePolicy(t, ctx, database, "environment", purpose,
				dimension, 900, actionThreshold[dimension], 3600, true)
		}
	}
	for _, invalid := range []struct {
		scope, purpose, dimension string
		window, threshold, block  int
	}{
		{"environment", "reauth", "identifier", 300, 5, 900},
		{"tenant", "login", "tenant_slug", 300, 30, 900},
		{"tenant", "password_reset", "selector", 900, 3, 3600},
		{"tenant", "login", "identifier", 301, 5, 900},
		{"tenant", "login", "identifier", 300, 4, 900},
		{"tenant", "login", "identifier", 300, 5, 901},
	} {
		assertIAMThrottlePolicy(t, ctx, database, invalid.scope, invalid.purpose,
			invalid.dimension, invalid.window, invalid.threshold, invalid.block, false)
	}
	var allowed bool
	err := database.runtime.QueryRowContext(ctx, iamThrottlePolicySQL,
		"tenant", "login", "identifier", 1, 300, 5, 900).Scan(&allowed)
	assertIAMSQLState(t, err, "42501")
}

func assertIAMThrottlePolicy(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	scope, purpose, dimension string, window, threshold, block int, want bool,
) {
	t.Helper()
	var allowed bool
	err := database.migrator.QueryRowContext(ctx, iamThrottlePolicySQL,
		scope, purpose, dimension, 1, window, threshold, block).Scan(&allowed)
	if err != nil || allowed != want {
		t.Fatalf("policy %s/%s/%s %d/%d/%d=%v want %v err=%v",
			scope, purpose, dimension, window, threshold, block, allowed, want, err)
	}
}
