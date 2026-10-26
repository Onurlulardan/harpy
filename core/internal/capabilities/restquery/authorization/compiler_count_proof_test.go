package authorization

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestCompilerRequiresCountSafeMandatoryPlanProof(t *testing.T) {
	spec, field := countUnsafeProofSpec(t)
	base := compilerBoundaryPlan(t, spec)
	visibilities, _ := base.Visibilities()
	requirement, _ := visibilities[0].RequirementSHA256()
	catalog, _ := visibilities[0].Catalog()
	path, _ := visibilities[0].Path()

	value, _ := resolved.NewScalarValue(api.LogicalTypeText, "blocked")
	predicate, err := resolved.NewMandatoryPredicate(
		field, api.OperatorEq, []resolved.Value{value},
	)
	if err != nil {
		t.Fatal(err)
	}
	filter, _ := resolved.NewMandatoryPredicateFilter(predicate)
	visibility, err := resolved.NewPredicateResourceVisibility(
		requirement, catalog, path, filter,
	)
	if err != nil {
		t.Fatal(err)
	}
	authorization, _ := base.AuthorizationContextSHA256()
	unsafe, err := resolved.MintAuthorizedPlan(
		spec, authorizedPlanTenant, authorizedPlanActor,
		authorization, []resolved.ResourceVisibility{visibility},
	)
	if err != nil {
		t.Fatal(err)
	}
	proved, proofErr := unsafe.ExactCountAuthorized()
	if proofErr != nil || proved {
		t.Fatal("count-unsafe mandatory predicate claimed exact-count proof")
	}
	postgresCompiler, _ := compiler.New(compilerBoundaryCodec(t))
	if _, err := postgresCompiler.CompileExactCount(unsafe); !errors.Is(
		err, compiler.ErrContract,
	) {
		t.Fatal("count-unsafe mandatory predicate minted exact-count SQL")
	}
}

func countUnsafeProofSpec(
	t *testing.T,
) (resolved.QuerySpec, api.FieldDefinition) {
	t.Helper()
	reference := authReference(t, "forms",
		"98000000-0000-4000-8000-000000000001",
		"98000000-0000-4000-8000-000000000002", "9")
	stable := authUUIDField(t,
		"98000000-0000-4000-8000-000000000003", "id", "proof_id")
	name := authTextField(t,
		"98000000-0000-4000-8000-000000000004", "name", "proof_name")
	profile, _ := api.NewTextProfile(
		api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
	)
	fieldType, _ := api.NewTextFieldType(profile)
	folded, _ := sqlidentifier.MintDynamic("fld_proof_unsafe_folded")
	unsafe := authField(t,
		"98000000-0000-4000-8000-000000000005", "unsafe", fieldType,
		"proof_unsafe", &folded, api.FieldCapabilitiesInput{
			Filters: []api.FilterCapabilityInput{
				{Operator: api.OperatorEq, CountSafe: false},
			},
		},
	)
	catalog := authCatalog(t, reference, "count_proof",
		[]api.FieldDefinition{stable, name, unsafe}, nil,
		"98000000-0000-4000-8000-000000000003",
		"98000000-0000-4000-8000-000000000004")
	catalogs, err := api.NewCatalogSet(reference, []api.FieldCatalog{catalog})
	if err != nil {
		t.Fatal(err)
	}
	request, _ := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"name"}, Page: snapshotPage(t), Count: api.CountModeExact,
	})
	spec, err := resolver.Resolve(request, catalogs)
	if err != nil {
		t.Fatal(err)
	}
	return spec, unsafe
}
