package authorization

import (
	"crypto/sha256"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/compiler"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func TestCompilerPublicBoundaryBindsPlanAndCursor(t *testing.T) {
	spec := snapshotRecordSpec(t, 1)
	plan := compilerBoundaryPlan(t, spec)
	codec := compilerBoundaryCodec(t)
	postgresCompiler, err := compiler.New(codec)
	if err != nil {
		t.Fatal(err)
	}
	first, err := postgresCompiler.CompileRecord(plan)
	if err != nil || first.ValidatePlan(plan) != nil {
		t.Fatalf("public compile rejected authorized plan: %v", err)
	}
	owner, ownerErr := first.Owner()
	class, classErr := first.Class()
	if ownerErr != nil || classErr != nil || owner != "forms" ||
		class != planner.StatementClassRecord {
		t.Fatal("statement owner did not come from the sealed root catalog")
	}
	token := compilerBoundaryCursor(t, codec, first)
	afterSpec := compilerBoundaryAfterSpec(t, spec, token)
	afterPlan := compilerBoundaryPlan(t, afterSpec)
	after, err := postgresCompiler.CompileRecord(afterPlan)
	if err != nil || after.ValidatePlan(afterPlan) != nil {
		t.Fatalf("plan-bound cursor did not compile: %v", err)
	}
	sql, sqlErr := after.SQL()
	arguments, argumentErr := after.Arguments()
	wantArguments := []any{
		"aaaaaaaa-bbbb-4ccc-8ddd-eeeeeeeeeeee", 26,
	}
	if sqlErr != nil || argumentErr != nil ||
		!strings.Contains(sql, `"fld_snapshot_forms_id" > $1`) ||
		!reflect.DeepEqual(arguments, wantArguments) {
		t.Fatalf("cursor SQL=%s args=%#v", sql, arguments)
	}
	assertCompilerRejectsSemanticDowngrade(t, postgresCompiler, spec)
	assertCompilerRootIncludeParity(t, postgresCompiler, first)
	assertCompilerRejectsCrossPlan(t, first, plan)
}

func assertCompilerRejectsSemanticDowngrade(
	t *testing.T, postgresCompiler compiler.Compiler, root resolved.QuerySpec,
) {
	t.Helper()
	unsupported := []resolved.AuthorizedPlan{
		compilerBoundaryPlan(t, compilerBoundaryExactSpec(t, root)),
		compilerBoundaryPlan(t,
			authorizationAggregateSpec(t, newAuthorizationQueryFixture(t))),
	}
	for _, plan := range unsupported {
		if _, err := postgresCompiler.CompileRecord(plan); !errors.Is(err, compiler.ErrContract) {
			t.Fatal("root compiler silently downgraded count/include/aggregate semantics")
		}
	}
}

func assertCompilerRootIncludeParity(
	t *testing.T, postgresCompiler compiler.Compiler, without compiler.RecordStatement,
) {
	t.Helper()
	includePlan := compilerBoundaryPlan(t, snapshotRecordSpec(t, 2))
	with, err := postgresCompiler.CompileRecord(includePlan)
	if err != nil || with.ValidatePlan(includePlan) != nil {
		t.Fatalf("root compiler rejected include-bearing plan: %v", err)
	}
	withoutSQL, _ := without.SQL()
	withSQL, _ := with.SQL()
	withoutArguments, _ := without.Arguments()
	withArguments, _ := with.Arguments()
	if withoutSQL != withSQL || !reflect.DeepEqual(withoutArguments, withArguments) {
		t.Fatal("include state changed root record SQL semantics")
	}
}

func assertCompilerRejectsCrossPlan(
	t *testing.T, statement compilerPlanValidator, plan resolved.AuthorizedPlan,
) {
	t.Helper()
	spec, _ := plan.Spec()
	visibilities, _ := plan.Visibilities()
	authorizationContext, _ := plan.AuthorizationContextSHA256()
	otherTenant, err := resolved.MintAuthorizedPlan(
		spec, uuid.MustParse("73000000-0000-4000-8000-000000000001"),
		authorizedPlanActor, authorizationContext, visibilities,
	)
	if err != nil || !errors.Is(statement.ValidatePlan(otherTenant), compiler.ErrContract) {
		t.Fatal("cross-tenant statement mix was accepted")
	}
	otherActor, err := resolved.MintAuthorizedPlan(
		spec, authorizedPlanTenant,
		uuid.MustParse("74000000-0000-4000-8000-000000000001"),
		authorizationContext, visibilities,
	)
	if err != nil || !errors.Is(statement.ValidatePlan(otherActor), compiler.ErrContract) {
		t.Fatal("cross-actor statement mix was accepted")
	}
	otherAuthorization, err := resolved.MintAuthorizedPlan(
		spec, authorizedPlanTenant, authorizedPlanActor,
		sha256.Sum256([]byte("different-authorization-context")), visibilities,
	)
	if err != nil || !errors.Is(
		statement.ValidatePlan(otherAuthorization), compiler.ErrContract,
	) {
		t.Fatal("cross-authorization-context statement mix was accepted")
	}
	requirement, _ := visibilities[0].RequirementSHA256()
	catalog, _ := visibilities[0].Catalog()
	path, _ := visibilities[0].Path()
	denied, err := resolved.NewNoResourceVisibility(requirement, catalog, path)
	if err != nil {
		t.Fatal(err)
	}
	otherMandatory, err := resolved.MintAuthorizedPlan(
		spec, authorizedPlanTenant, authorizedPlanActor,
		authorizationContext, []resolved.ResourceVisibility{denied},
	)
	if err != nil || !errors.Is(statement.ValidatePlan(otherMandatory), compiler.ErrContract) {
		t.Fatal("cross-mandatory-visibility statement mix was accepted")
	}
	otherSpec := authorizationRecordSpec(t, newAuthorizationQueryFixture(t))
	otherCatalog := compilerBoundaryPlan(t, otherSpec)
	if !errors.Is(statement.ValidatePlan(otherCatalog), compiler.ErrContract) {
		t.Fatal("cross-catalog statement mix was accepted")
	}
}

type compilerPlanValidator interface {
	ValidatePlan(resolved.AuthorizedPlan) error
}

func compilerBoundaryPlan(
	t *testing.T, spec resolved.QuerySpec,
) resolved.AuthorizedPlan {
	t.Helper()
	set, evaluations := allowAllPlanFixture(t, spec)
	plan, err := buildAuthorizedPlan(
		spec, set, authorizedPlanTenant, authorizedPlanActor, evaluations,
	)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}
