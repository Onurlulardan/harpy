package iam

import (
	"context"
	"errors"
	"slices"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type iamSnapshotHTTPTestPlatform struct {
	*identityHTTPTestPlatform
}

func (*iamSnapshotHTTPTestPlatform) ResolveActiveTenantIDBySlug(
	context.Context, string,
) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (*iamSnapshotHTTPTestPlatform) ResolveEnvironmentID(
	context.Context,
) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func TestIAMSnapshotSurfacesRequireAtomicIAMLayoutGeneration(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	ctx := context.Background()
	tests := []struct {
		name        string
		unavailable error
		call        func(*iamSnapshotHTTPTestPlatform) (platformtx.ScopeRequest, error)
	}{
		{
			name: "invitation", unavailable: ErrInvitationHTTPUnavailable,
			call: func(platform *iamSnapshotHTTPTestPlatform) (platformtx.ScopeRequest, error) {
				return (&invitationHTTPApplication{pool: platform}).snapshotScope(ctx, fixture.claims)
			},
		},
		{
			name: "lifecycle", unavailable: ErrLifecycleHTTPUnavailable,
			call: func(platform *iamSnapshotHTTPTestPlatform) (platformtx.ScopeRequest, error) {
				return (&lifecycleHTTPApplication{pool: platform}).snapshotScope(ctx, fixture.claims)
			},
		},
		{
			name: "permission_request", unavailable: ErrPermissionRequestHTTPUnavailable,
			call: func(platform *iamSnapshotHTTPTestPlatform) (platformtx.ScopeRequest, error) {
				return (&permissionRequestHTTPApplication{pool: platform}).snapshotScopeRequest(
					ctx, fixture.claims,
				)
			},
		},
		{
			name: "rbac", unavailable: ErrRBACHTTPUnavailable,
			call: func(platform *iamSnapshotHTTPTestPlatform) (platformtx.ScopeRequest, error) {
				return (&rbacHTTPApplication{pool: platform}).snapshotScopeRequest(ctx, fixture.claims)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			platform := iamSnapshotPlatform(map[string]int64{
				"iam": 9, "system_resource_layout": 9,
			})
			request, err := test.call(platform)
			if err != nil || request.Mode != platformtx.SnapshotQuery ||
				!slices.Equal(platform.domains,
					[]string{"iam", "system_resource_layout"}) {
				t.Fatalf("request=%+v domains=%v error=%v",
					request, platform.domains, err)
			}
			for name, generations := range map[string]map[string]int64{
				"single domain": {"iam": 9},
				"unequal":       {"iam": 9, "system_resource_layout": 10},
			} {
				t.Run(name, func(t *testing.T) {
					_, err := test.call(iamSnapshotPlatform(generations))
					if !errors.Is(err, test.unavailable) {
						t.Fatalf("error=%v", err)
					}
				})
			}
		})
	}
}

func iamSnapshotPlatform(generations map[string]int64) *iamSnapshotHTTPTestPlatform {
	return &iamSnapshotHTTPTestPlatform{identityHTTPTestPlatform: &identityHTTPTestPlatform{
		generations: generations,
	}}
}
