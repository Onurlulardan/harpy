package executor

import (
	"context"
	"errors"
	"fmt"
	"testing"

	v1query "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

type executorV1ManifestSource struct{}

func (executorV1ManifestSource) VerifiedV1QueryManifest() (
	v1query.ManifestSnapshot, error,
) {
	return v1query.ManifestSnapshot{
		ManifestID:  "51000000-0000-4000-8000-000000000099",
		OwnerModule: "forms", Domain: "forms",
		DefaultLimit: 2, MaximumPageSize: 500,
		DefaultSortField: "id", DefaultSortDirection: "asc",
		FilterFields: map[string]v1query.FilterFieldBinding{},
		SortFields: map[string]v1query.SortFieldBinding{
			"id": {Name: "id"},
		},
		OperatorAliases: map[string]v1query.OperatorBinding{},
		QueryKeyAliases: map[string]v1query.QueryKeyBinding{},
	}, nil
}

func executorV1PagePlan(
	t *testing.T, plan resolved.AuthorizedPlan, page int64, limit int,
) v1query.PagePlan {
	t.Helper()
	coordinator, err := v1query.NewCoordinator(
		executorV1ManifestSource{}, staticV1Authorizer{plan: plan}, rejectV1Evidence{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := coordinator.AdmitManifest(context.Background()); err != nil {
		t.Fatal(err)
	}
	result, err := coordinator.Accept(
		context.Background(), fmt.Sprintf("page=%d&limit=%d", page, limit),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, pagePlan, err := result.ConsumeExecution()
	if err != nil {
		t.Fatal(err)
	}
	return pagePlan
}

type staticV1Authorizer struct{ plan resolved.AuthorizedPlan }

func (port staticV1Authorizer) ResolveAndAuthorizeV1(
	_ context.Context, _ api.QueryRequest,
) (resolved.AuthorizedPlan, error) {
	return port.plan, nil
}

type rejectV1Evidence struct{}

func (rejectV1Evidence) AdmitManifest(
	context.Context, contractretirement.Batch,
) (contractretirement.Decision, error) {
	return contractretirement.Decision{}, errors.New("unexpected V1 evidence call")
}

func (rejectV1Evidence) RecordHits(
	context.Context, contractretirement.Batch,
) (contractretirement.Decision, error) {
	return contractretirement.Decision{}, errors.New("unexpected V1 evidence call")
}
