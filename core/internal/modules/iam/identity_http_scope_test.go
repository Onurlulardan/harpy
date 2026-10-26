package iam

import (
	"context"
	"errors"
	"slices"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type identityHTTPTestPlatform struct {
	generations map[string]int64
	err         error
	tenant      uuid.UUID
	domains     []string
}

func (platform *identityHTTPTestPlatform) ResolveSnapshotDomainGenerations(
	_ context.Context, tenant uuid.UUID, domains []string,
) (map[string]int64, error) {
	platform.tenant = tenant
	platform.domains = append([]string(nil), domains...)
	return platform.generations, platform.err
}

func TestIdentityCommandScopeRequiresAtomicIAMLayoutGeneration(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	platform := &identityHTTPTestPlatform{generations: map[string]int64{
		"iam": 7, "system_resource_layout": 7,
	}}
	application := &identityHTTPApplication{pool: platform}
	request, err := application.commandScopeRequest(
		context.Background(), fixture.claims, "iam.identity.user.create",
	)
	if err != nil || request.Mode != platformtx.Command ||
		platform.tenant != fixture.claims.Tenant() ||
		!slices.Equal(platform.domains, []string{"iam", "system_resource_layout"}) {
		t.Fatalf("request=%+v tenant=%s domains=%v error=%v",
			request, platform.tenant, platform.domains, err)
	}
	for name, generations := range map[string]map[string]int64{
		"missing layout": {"iam": 7},
		"mismatch":       {"iam": 7, "system_resource_layout": 8},
		"zero":           {"iam": 0, "system_resource_layout": 0},
		"extra":          {"iam": 7, "system_resource_layout": 7, "forms": 7},
	} {
		t.Run(name, func(t *testing.T) {
			application.pool = &identityHTTPTestPlatform{generations: generations}
			_, err := application.commandScopeRequest(
				context.Background(), fixture.claims, "iam.identity.user.create",
			)
			if !errors.Is(err, ErrIdentityHTTPUnavailable) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestIdentitySnapshotScopeRequiresAtomicIAMLayoutGeneration(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	platform := &identityHTTPTestPlatform{generations: map[string]int64{
		"iam": 9, "system_resource_layout": 9,
	}}
	application := &identityHTTPApplication{pool: platform}
	request, err := application.snapshotScopeRequest(
		context.Background(), fixture.claims, "restquery.execute",
	)
	if err != nil || request.Mode != platformtx.SnapshotQuery ||
		!slices.Equal(platform.domains, []string{"iam", "system_resource_layout"}) {
		t.Fatalf("request=%+v domains=%v error=%v", request, platform.domains, err)
	}
	for name, generations := range map[string]map[string]int64{
		"single domain": {"iam": 9},
		"unequal":       {"iam": 9, "system_resource_layout": 10},
	} {
		t.Run(name, func(t *testing.T) {
			application.pool = &identityHTTPTestPlatform{generations: generations}
			if _, err := application.snapshotScopeRequest(
				context.Background(), fixture.claims, "restquery.execute",
			); !errors.Is(err, ErrIdentityHTTPUnavailable) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestIdentityResourceAuthorizationKeysAreClosed(t *testing.T) {
	for resource, want := range map[IdentityResource]string{
		IdentityUsers: "iam.users", IdentityRoles: "iam.roles",
		IdentityOrganizations: "iam.organizations",
	} {
		actual, err := identityResourceKey(resource)
		if err != nil || actual != want {
			t.Fatalf("resource=%q key=%q error=%v", resource, actual, err)
		}
	}
	if _, err := identityResourceKey("forms"); err == nil {
		t.Fatal("unknown resource received an authorization key")
	}
}
