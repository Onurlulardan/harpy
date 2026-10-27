//go:build integration

package integrationtest

import (
	"context"
	"net/url"
	"testing"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const catalogActorID = "018f2000-0000-7000-8000-000000000010"

func openCatalogCoordinator(
	t *testing.T, ctx context.Context, database catalogPG18,
) *platformtx.Coordinator {
	t.Helper()
	parsed, err := url.Parse(database.dsn)
	if err != nil {
		t.Fatal(err)
	}
	parsed.User = url.UserPassword("harpy_v2_runtime", "runtime-only")
	pool, err := platformpostgres.OpenIntegrationAPI(
		ctx, t, parsed.String(), catalogExtensionDigest,
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	coordinator, err := platformtx.NewCoordinator(pool)
	if err != nil {
		t.Fatal(err)
	}
	return coordinator
}

func catalogLifecycleRequest(t *testing.T, tenantID, operation string) platformtx.ScopeRequest {
	return catalogLifecycleRequestForActor(t, tenantID, catalogActorID, operation)
}

func catalogLifecycleRequestForActor(
	t *testing.T, tenantID, actorID, operation string,
) platformtx.ScopeRequest {
	t.Helper()
	generation := catalogTenantGeneration(t, tenantID)
	groups := make([]platformtx.CommandAuthorityGroup, 0, len(catalogAuthorityGroups))
	domains := make([]platformtx.CommandAuthorityDomain, 0, 7)
	for _, group := range catalogAuthorityGroups {
		if group.id == "documents" && operation != "catalog.operation.resume" {
			continue
		}
		groups = append(groups, platformtx.CommandAuthorityGroup{
			ID: group.id, Generation: generation,
		})
		for _, domain := range group.members {
			domains = append(domains, platformtx.CommandAuthorityDomain{
				Name: domain, Generation: generation,
			})
		}
	}
	participants := []string{"analytics", "catalog", "forms", "iam", "workflow"}
	if operation == "catalog.operation.resume" {
		participants = []string{
			"analytics", "catalog", "documents", "forms", "iam", "workflow",
		}
	}
	request, err := platformtx.AdmitMultiGroupCommandScope(
		uuid.MustParse(actorID), uuid.MustParse(tenantID), operation,
		groups, domains, participants,
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func catalogAdoptionRequest(t *testing.T, tenantID string) platformtx.ScopeRequest {
	t.Helper()
	generation := catalogTenantGeneration(t, tenantID)
	request, err := platformtx.AdmitCommandScope(
		uuid.MustParse(catalogActorID), uuid.MustParse(tenantID),
		"catalog.adoption.verify", "catalog", generation,
		map[string]int64{"catalog": generation},
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func catalogTenantGeneration(t *testing.T, tenantID string) int64 {
	t.Helper()
	for _, tenant := range catalogTenants {
		if tenant.id == tenantID {
			return tenant.generation
		}
	}
	t.Fatalf("unknown Catalog tenant %s", tenantID)
	return 0
}
