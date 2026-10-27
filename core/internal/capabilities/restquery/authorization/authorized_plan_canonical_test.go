package authorization

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestMandatoryVisibilityCanonicalizesSetValueOrder(t *testing.T) {
	fixture := newAuthorizationQueryFixture(t)
	spec := authorizationRecordSpec(t, fixture)
	set, evaluations := allowAllPlanFixture(t, spec)
	forward := mandatoryValuesFilter(
		t, authRootName, api.LogicalTypeText, api.OperatorIn, "Ada", "Grace",
	)
	reverse := mandatoryValuesFilter(
		t, authRootName, api.LogicalTypeText, api.OperatorIn, "Grace", "Ada",
	)
	forwardCanonical := mandatoryCanonicalForFilter(
		t, spec, set, evaluations, forward,
	)
	reverseCanonical := mandatoryCanonicalForFilter(
		t, spec, set, evaluations, reverse,
	)
	if !bytes.Equal(forwardCanonical, reverseCanonical) {
		t.Fatal("set-valued mandatory operands were not canonicalized")
	}

	lowerFirst := mandatoryValuesFilter(
		t, authRootAmount, api.LogicalTypeDecimal, api.OperatorBetween,
		"1.00", "2.00",
	)
	upperFirst := mandatoryValuesFilter(
		t, authRootAmount, api.LogicalTypeDecimal, api.OperatorBetween,
		"2.00", "1.00",
	)
	if bytes.Equal(
		mandatoryCanonicalForFilter(t, spec, set, evaluations, lowerFirst),
		mandatoryCanonicalForFilter(t, spec, set, evaluations, upperFirst),
	) {
		t.Fatal("ordered mandatory pair operands were treated as a set")
	}
}

func TestMandatoryVisibilityCanonicalHardBound(t *testing.T) {
	fixture := newAuthorizationQueryFixture(t)
	spec := authorizationRecordSpec(t, fixture)
	set, evaluations := allowAllPlanFixture(t, spec)
	base := mandatoryValuesFilter(
		t, authRootName, api.LogicalTypeText, api.OperatorEq, "",
	)
	baseCanonical := mandatoryCanonicalForFilter(t, spec, set, evaluations, base)
	const hardLimit = 64 * 1024
	valueSize := hardLimit - len(baseCanonical)
	if valueSize < 1 {
		t.Fatal("mandatory canonical fixture already exceeds the hard bound")
	}
	exact := mandatoryValuesFilter(
		t, authRootName, api.LogicalTypeText, api.OperatorEq,
		strings.Repeat("x", valueSize),
	)
	canonical := mandatoryCanonicalForFilter(t, spec, set, evaluations, exact)
	if len(canonical) != hardLimit {
		t.Fatalf("exact mandatory canonical size = %d", len(canonical))
	}
	overflow := mandatoryValuesFilter(
		t, authRootName, api.LogicalTypeText, api.OperatorEq,
		strings.Repeat("x", valueSize+1),
	)
	if _, err := buildAuthorizedPlan(
		spec, set, authorizedPlanTenant, authorizedPlanActor,
		evaluationsWithRootFilter(t, evaluations, overflow),
	); err == nil {
		t.Fatal("mandatory canonical overflow minted a plan")
	}
}

func mandatoryValuesFilter(
	t *testing.T,
	fieldID string,
	kind api.LogicalType,
	operator api.Operator,
	values ...string,
) MandatoryFilter {
	t.Helper()
	resolvedValues := make([]resolved.Value, len(values))
	for index, value := range values {
		resolvedValues[index] = resolvedScalar(t, kind, value)
	}
	predicate, err := NewMandatoryPredicate(
		fieldID, kind, operator, resolvedValues,
	)
	if err != nil {
		t.Fatal(err)
	}
	filter, err := NewMandatoryPredicateFilter(predicate)
	if err != nil {
		t.Fatal(err)
	}
	return filter
}

func mandatoryCanonicalForFilter(
	t *testing.T,
	spec resolved.QuerySpec,
	set requirementSet,
	evaluations []policyEvaluation,
	filter MandatoryFilter,
) []byte {
	t.Helper()
	plan, err := buildAuthorizedPlan(
		spec, set, authorizedPlanTenant, authorizedPlanActor,
		evaluationsWithRootFilter(t, evaluations, filter),
	)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := plan.MandatoryCanonical()
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}
