package resolver

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestResolveListArityDeduplicatesAfterNormalization(t *testing.T) {
	t.Parallel()
	fieldType := mustTextType(t, api.TextCasefoldDisabled, api.TextAccentDisabled)
	values, err := ResolveLiterals(fieldType, api.OperatorIn, []api.Literal{
		mustTextLiteral(t, "Cafe\u0301"),
		mustJSONLiteral(t, `"Café"`),
		mustTextLiteral(t, "OTHER"),
		mustTextLiteral(t, "OTHER"),
	})
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, "Café", "OTHER")
}

func TestResolveCollectionElementsDeduplicateInFirstOccurrenceOrder(t *testing.T) {
	t.Parallel()
	profile, _ := api.NewTextProfile(api.TextCasefoldDisabled, api.TextAccentDisabled)
	fieldType, _ := api.NewTextListFieldType(profile)
	values, err := ResolveLiterals(fieldType, api.OperatorContainsAll, []api.Literal{
		mustTextLiteral(t, "B"), mustTextLiteral(t, "A"), mustTextLiteral(t, "B"),
	})
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, "B", "A")
}

func TestResolveBooleanListDeduplicatesCanonicalValues(t *testing.T) {
	t.Parallel()
	fieldType, _ := api.NewBooleanFieldType()
	values, err := ResolveLiterals(fieldType, api.OperatorIn, []api.Literal{
		mustJSONLiteral(t, `true`), mustTextLiteral(t, "true"), mustJSONLiteral(t, `false`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(values) != 2 {
		t.Fatalf("deduplicated boolean count = %d", len(values))
	}
	first, _, _ := values[0].Boolean()
	second, _, _ := values[1].Boolean()
	if !first || second {
		t.Fatalf("boolean order = %v, %v", first, second)
	}
}

func TestResolveBetweenDoesNotDeduplicatePair(t *testing.T) {
	t.Parallel()
	profile, _ := api.NewIntegerProfile(0, 100)
	fieldType, _ := api.NewIntegerFieldType(profile)
	values, err := ResolveLiterals(fieldType, api.OperatorBetween, []api.Literal{
		mustTextLiteral(t, "10"), mustJSONLiteral(t, `"10"`),
	})
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, "10", "10")
}
