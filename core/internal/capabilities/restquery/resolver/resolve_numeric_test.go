package resolver

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestResolveIntegerSupportsPOSTAndGETCanonicalStrings(t *testing.T) {
	t.Parallel()
	profile, _ := api.NewIntegerProfile(-100, 100)
	fieldType, _ := api.NewIntegerFieldType(profile)
	for _, literal := range []api.Literal{
		mustJSONLiteral(t, `"42"`), mustTextLiteral(t, "42"),
	} {
		values, err := ResolveLiterals(fieldType, api.OperatorEq, []api.Literal{literal})
		if err != nil {
			t.Fatal(err)
		}
		requireScalars(t, values, "42")
	}
	_, err := ResolveLiterals(
		fieldType, api.OperatorEq, []api.Literal{mustJSONLiteral(t, `42`)},
	)
	requireResolutionCode(t, err, restqueryv1.IssueTypeMismatch)
}

func TestResolveIntegerSeparatesCanonicalAndRangeFailures(t *testing.T) {
	t.Parallel()
	profile, _ := api.NewIntegerProfile(-10, 10)
	fieldType, _ := api.NewIntegerFieldType(profile)
	for _, literal := range []string{"01", "-0", "+1", "1.0"} {
		_, err := ResolveLiterals(
			fieldType, api.OperatorEq, []api.Literal{mustTextLiteral(t, literal)},
		)
		requireResolutionCode(t, err, restqueryv1.IssueValueNotCanonical)
	}
	for _, literal := range []string{"11", "9223372036854775808"} {
		_, err := ResolveLiterals(
			fieldType, api.OperatorEq, []api.Literal{mustTextLiteral(t, literal)},
		)
		requireResolutionCode(t, err, restqueryv1.IssueValueOutOfRange)
	}
}

func TestResolveDecimalPadsDeclaredScale(t *testing.T) {
	t.Parallel()
	profile, _ := api.NewDecimalProfile(5, 2)
	fieldType, _ := api.NewDecimalFieldType(profile)
	for _, test := range []struct {
		literal api.Literal
		want    string
	}{
		{mustJSONLiteral(t, `"1.2"`), "1.20"},
		{mustTextLiteral(t, "0"), "0.00"},
		{mustTextLiteral(t, "-0.01"), "-0.01"},
	} {
		values, err := ResolveLiterals(fieldType, api.OperatorEq, []api.Literal{test.literal})
		if err != nil {
			t.Fatal(err)
		}
		requireScalars(t, values, test.want)
	}
}

func TestResolveDecimalSeparatesTypeCanonicalAndPrecisionFailures(t *testing.T) {
	t.Parallel()
	profile, _ := api.NewDecimalProfile(4, 2)
	fieldType, _ := api.NewDecimalFieldType(profile)
	_, err := ResolveLiterals(
		fieldType, api.OperatorEq, []api.Literal{mustJSONLiteral(t, `1.2`)},
	)
	requireResolutionCode(t, err, restqueryv1.IssueTypeMismatch)
	for _, literal := range []string{"1e2", "01.2", "1.", "-0.00"} {
		_, err = ResolveLiterals(
			fieldType, api.OperatorEq, []api.Literal{mustTextLiteral(t, literal)},
		)
		requireResolutionCode(t, err, restqueryv1.IssueValueNotCanonical)
	}
	for _, literal := range []string{"123.45", "1.234"} {
		_, err = ResolveLiterals(
			fieldType, api.OperatorEq, []api.Literal{mustTextLiteral(t, literal)},
		)
		requireResolutionCode(t, err, restqueryv1.IssueValueOutOfRange)
	}
}

func TestResolveDecimalAllowsZeroWhenPrecisionEqualsScale(t *testing.T) {
	t.Parallel()
	profile, _ := api.NewDecimalProfile(2, 2)
	fieldType, _ := api.NewDecimalFieldType(profile)
	values, err := ResolveLiterals(
		fieldType, api.OperatorEq, []api.Literal{mustTextLiteral(t, "0.12")},
	)
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, "0.12")
}
