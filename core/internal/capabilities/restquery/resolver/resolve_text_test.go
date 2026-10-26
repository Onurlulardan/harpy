package resolver

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestResolveTextNormalizesPOSTAndGETToSameExactValue(t *testing.T) {
	t.Parallel()
	fieldType := mustTextType(t, api.TextCasefoldDisabled, api.TextAccentDisabled)
	post := mustJSONLiteral(t, `"Cafe\u0301"`)
	get := mustTextLiteral(t, "Cafe\u0301")
	for _, literal := range []api.Literal{post, get} {
		values, err := ResolveLiterals(fieldType, api.OperatorEq, []api.Literal{literal})
		if err != nil {
			t.Fatal(err)
		}
		requireScalars(t, values, "Café")
	}
}

func TestResolvePatternAppliesFoldAndAccentProfile(t *testing.T) {
	t.Parallel()
	folded := mustTextType(t, api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled)
	values, err := ResolveLiterals(
		folded, api.OperatorContains, []api.Literal{mustTextLiteral(t, "Straße")},
	)
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, "strasse")

	accent := mustTextType(
		t, api.TextCasefoldUnicodeDefaultV1, api.TextAccentInsensitiveV1,
	)
	values, err = ResolveLiterals(
		accent, api.OperatorStartsWith, []api.Literal{mustJSONLiteral(t, `"CAFÉ"`)},
	)
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, "cafe")
}

func TestResolveExactTextNeverUsesCompanionNormalization(t *testing.T) {
	t.Parallel()
	fieldType := mustTextType(
		t, api.TextCasefoldUnicodeDefaultV1, api.TextAccentInsensitiveV1,
	)
	values, err := ResolveLiterals(
		fieldType, api.OperatorEq, []api.Literal{mustTextLiteral(t, "Straße CAFÉ")},
	)
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, "Straße CAFÉ")
}

func TestResolvePatternPreservesLiteralMetacharacters(t *testing.T) {
	t.Parallel()
	fieldType := mustTextType(
		t, api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
	)
	values, err := ResolveLiterals(
		fieldType, api.OperatorContains, []api.Literal{mustTextLiteral(t, `A%_\B`)},
	)
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, `a%_\b`)
}

func TestAccentProfileRemovesEveryUnicodeMarkCategory(t *testing.T) {
	t.Parallel()
	fieldType := mustTextType(
		t, api.TextCasefoldUnicodeDefaultV1, api.TextAccentInsensitiveV1,
	)
	values, err := ResolveLiterals(
		fieldType, api.OperatorContains,
		[]api.Literal{mustTextLiteral(t, "a\u0301a\u0903a\u20dd")},
	)
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, "aaa")
}

func TestResolveTextDistinguishesTypeCanonicalAndPolicyErrors(t *testing.T) {
	t.Parallel()
	exact := mustTextType(t, api.TextCasefoldDisabled, api.TextAccentDisabled)
	_, err := ResolveLiterals(
		exact, api.OperatorEq, []api.Literal{mustJSONLiteral(t, `true`)},
	)
	requireResolutionCode(t, err, restqueryv1.IssueTypeMismatch)

	_, err = ResolveLiterals(
		exact, api.OperatorEq, []api.Literal{mustJSONLiteral(t, `"\u0000"`)},
	)
	requireResolutionCode(t, err, restqueryv1.IssueNULForbidden)

	_, err = ResolveLiterals(
		exact, api.OperatorContains, []api.Literal{mustTextLiteral(t, "value")},
	)
	requireResolutionCode(t, err, restqueryv1.IssueOperatorNotAllowed)
}

func TestResolvePatternRejectsEmptyEffectiveValue(t *testing.T) {
	t.Parallel()
	accent := mustTextType(
		t, api.TextCasefoldUnicodeDefaultV1, api.TextAccentInsensitiveV1,
	)
	_, err := ResolveLiterals(
		accent, api.OperatorContains, []api.Literal{mustTextLiteral(t, "\u0301")},
	)
	requireResolutionCode(t, err, restqueryv1.IssueValueNotAllowed)

	exact := mustTextType(t, api.TextCasefoldDisabled, api.TextAccentDisabled)
	values, err := ResolveLiterals(
		exact, api.OperatorEq, []api.Literal{mustTextLiteral(t, "")},
	)
	if err != nil {
		t.Fatal(err)
	}
	requireScalars(t, values, "")
}
