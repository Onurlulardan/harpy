//go:build integration

package executor

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	v1query "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/google/uuid"
)

type pg18V1ManifestSource struct{}

func (pg18V1ManifestSource) VerifiedV1QueryManifest() (
	v1query.ManifestSnapshot, error,
) {
	return v1query.ManifestSnapshot{
		ManifestID:  "51000000-0000-4000-8000-000000000098",
		OwnerModule: "forms", Domain: "forms",
		DefaultLimit: 2, MaximumPageSize: 500,
		DefaultSortField: "name", DefaultSortDirection: "asc",
		FilterFields: map[string]v1query.FilterFieldBinding{
			"name": {Name: "name", Shape: v1query.FilterShapeScalar},
		},
		SortFields: map[string]v1query.SortFieldBinding{
			"name": {Name: "name"}, "id": {Name: "id"},
		},
		OperatorAliases: map[string]v1query.OperatorBinding{},
		QueryKeyAliases: map[string]v1query.QueryKeyBinding{},
	}, nil
}

func pg18V1Plan(
	t *testing.T, raw string, tenant string,
) v1query.Accepted {
	t.Helper()
	coordinator, err := v1query.NewCoordinator(
		pg18V1ManifestSource{},
		pg18V1AuthorizationPort{testingT: t, tenant: tenant},
		rejectV1Evidence{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := coordinator.AdmitManifest(context.Background()); err != nil {
		t.Fatal(err)
	}
	accepted, err := coordinator.Accept(context.Background(), raw)
	if err != nil {
		t.Fatal(err)
	}
	return accepted
}

type pg18V1AuthorizationPort struct {
	testingT *testing.T
	tenant   string
}

func (port pg18V1AuthorizationPort) ResolveAndAuthorizeV1(
	_ context.Context, request api.QueryRequest,
) (resolved.AuthorizedPlan, error) {
	t := port.testingT
	spec, err := resolver.Resolve(request, pg18CatalogSet(t, "forms"))
	if err != nil {
		return resolved.AuthorizedPlan{}, err
	}
	reference, _ := spec.Reference()
	path, _ := spec.RootPath()
	visibility, err := resolved.NewAllResourceVisibility(
		sha256.Sum256([]byte("pg18-v1-restquery-requirement")), reference, path,
	)
	if err != nil {
		return resolved.AuthorizedPlan{}, err
	}
	plan, err := resolved.MintAuthorizedPlan(
		spec, uuid.MustParse(port.tenant), uuid.MustParse(pg18Actor),
		sha256.Sum256([]byte("pg18-v1-restquery-authorization")),
		[]resolved.ResourceVisibility{visibility},
	)
	if err != nil {
		return resolved.AuthorizedPlan{}, err
	}
	if request.Count() != api.CountModeExact {
		return resolved.AuthorizedPlan{}, fmt.Errorf("V1 request lost exact count")
	}
	return plan, nil
}
