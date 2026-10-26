//go:build integration

package integrationtest

import (
	"context"
	"testing"
	"time"
)

func TestProfileFencedRecoveryAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 7*time.Minute)
			defer cancel()
			scenario := prepareRecoveryPG18Scenario(t, ctx, origin)
			assertRecoveryUserRootFence(t, ctx, scenario)
			assertRecoveryHistoricalEmptyMapping(t, ctx, scenario)
			assertRecoveryFutureRejectedBeforeWrites(t, ctx, scenario)
		})
	}
}
