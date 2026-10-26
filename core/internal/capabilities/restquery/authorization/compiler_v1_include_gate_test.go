package authorization

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	v1query "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

type compilerV1ManifestSource struct{}

func (compilerV1ManifestSource) VerifiedV1QueryManifest() (
	v1query.ManifestSnapshot, error,
) {
	return v1query.ManifestSnapshot{
		ManifestID:  "51000000-0000-4000-8000-000000000098",
		OwnerModule: "forms", Domain: "forms",
		DefaultLimit: 25, MaximumPageSize: 500,
		DefaultSortField: "id", DefaultSortDirection: "asc",
		FilterFields: map[string]v1query.FilterFieldBinding{},
		SortFields: map[string]v1query.SortFieldBinding{
			"id": {Name: "id"},
		},
		OperatorAliases: map[string]v1query.OperatorBinding{},
		QueryKeyAliases: map[string]v1query.QueryKeyBinding{},
	}, nil
}

type compilerV1Authorizer struct{ plan resolved.AuthorizedPlan }

func (port compilerV1Authorizer) ResolveAndAuthorizeV1(
	_ context.Context, _ api.QueryRequest,
) (resolved.AuthorizedPlan, error) {
	return port.plan, nil
}

type compilerV1Evidence struct{}

func (compilerV1Evidence) AdmitManifest(
	context.Context, contractretirement.Batch,
) (contractretirement.Decision, error) {
	return contractretirement.Decision{}, errors.New("unexpected V1 evidence")
}

func (compilerV1Evidence) RecordHits(
	context.Context, contractretirement.Batch,
) (contractretirement.Decision, error) {
	return contractretirement.Decision{}, errors.New("unexpected V1 evidence")
}

func TestV1CompilerRejectsIncludeBearingPlan(t *testing.T) {
	root := snapshotRecordSpec(t, 2)
	without := compilerBoundaryPlan(t, compilerBoundaryExactSpec(t, root))
	with := compilerBoundaryPlan(t, compilerBoundaryExactIncludeSpec(t, root))
	postgresCompiler, _ := compiler.New(compilerBoundaryCodec(t))
	ordinal := compilerV1Ordinal(t, without)
	if _, err := postgresCompiler.CompileV1Predecessor(
		with, ordinal,
	); !errors.Is(err, compiler.ErrContract) {
		t.Fatal("V1 predecessor accepted include-bearing plan")
	}
	predecessor, err := postgresCompiler.CompileV1Predecessor(without, ordinal)
	if err != nil {
		t.Fatal(err)
	}
	identifier, _ := resolved.NewScalarValue(
		api.LogicalTypeUUID, "72000000-0000-4000-8000-000000000001",
	)
	identifierLast, _ := cursor.NewLastValue(identifier)
	tuple, err := predecessor.MintTuple([]cursor.LastValue{identifierLast})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := postgresCompiler.CompileV1Page(
		without, ordinal, predecessor, tuple,
	); err != nil {
		t.Fatal("valid include-free V1 page was rejected")
	}
	if _, err := postgresCompiler.CompileV1Page(
		with, ordinal, predecessor, tuple,
	); !errors.Is(err, compiler.ErrContract) {
		t.Fatal("V1 page accepted include-bearing plan")
	}
}

func TestV1AcceptedRejectsIncludeBearingPlanForEveryPage(t *testing.T) {
	root := snapshotRecordSpec(t, 2)
	with := compilerBoundaryPlan(t, compilerBoundaryExactIncludeSpec(t, root))
	coordinator, err := v1query.NewCoordinator(
		compilerV1ManifestSource{}, compilerV1Authorizer{plan: with},
		compilerV1Evidence{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := coordinator.AdmitManifest(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, page := range []int{1, 2} {
		if _, err := coordinator.Accept(
			context.Background(), fmt.Sprintf("page=%d&limit=%d", page, 25),
		); err == nil {
			t.Fatalf("V1 page %d accepted include-bearing authorized plan", page)
		}
	}
}

func compilerV1Ordinal(
	t *testing.T, plan resolved.AuthorizedPlan,
) v1query.PredecessorOrdinal {
	t.Helper()
	coordinator, err := v1query.NewCoordinator(
		compilerV1ManifestSource{}, compilerV1Authorizer{plan: plan},
		compilerV1Evidence{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := coordinator.AdmitManifest(context.Background()); err != nil {
		t.Fatal(err)
	}
	accepted, err := coordinator.Accept(
		context.Background(), fmt.Sprintf("page=%d&limit=%d", 2, 25),
	)
	if err != nil {
		t.Fatal(err)
	}
	_, page, err := accepted.ConsumeExecution()
	if err != nil {
		t.Fatal(err)
	}
	ordinal, present, err := page.PredecessorOrdinal()
	if err != nil || !present {
		t.Fatal("V1 predecessor ordinal missing")
	}
	return ordinal
}
