package iam

import (
	"context"
	"errors"
	"slices"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestProfileSnapshotScopeRequiresAtomicIAMLayoutGeneration(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	platform := &identityHTTPTestPlatform{generations: map[string]int64{
		"iam": 9, "system_resource_layout": 9,
	}}
	application := &profileHTTPApplication{pool: platform}
	request, err := application.snapshotScopeRequest(
		context.Background(), fixture.claims,
	)
	if err != nil || request.Mode != platformtx.SnapshotQuery ||
		platform.tenant != fixture.claims.Tenant() ||
		!slices.Equal(platform.domains, []string{"iam", "system_resource_layout"}) {
		t.Fatalf("request=%+v tenant=%s domains=%v error=%v",
			request, platform.tenant, platform.domains, err)
	}
	for name, generations := range map[string]map[string]int64{
		"single domain": {"iam": 9},
		"unequal":       {"iam": 9, "system_resource_layout": 10},
	} {
		t.Run(name, func(t *testing.T) {
			application.pool = &identityHTTPTestPlatform{generations: generations}
			_, err := application.snapshotScopeRequest(
				context.Background(), fixture.claims,
			)
			if !errors.Is(err, ErrProfileHTTPUnavailable) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
