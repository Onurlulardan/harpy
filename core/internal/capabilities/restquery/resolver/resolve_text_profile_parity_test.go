package resolver

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/textprofile"
)

func TestResolveTextMatchesPlatformProfileAuthority(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		casefold  api.TextCasefoldProfile
		accent    api.TextAccentProfile
		operator  api.Operator
		input     string
		normalize func(string) string
	}{
		{"exact", api.TextCasefoldDisabled, api.TextAccentDisabled,
			api.OperatorEq, "Cafe\u0301", textprofile.ExactNFC},
		{"casefold", api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
			api.OperatorContains, "Stra\u00dfe", textprofile.UnicodeDefaultCasefoldV1},
		{"Turkish I", api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
			api.OperatorContains, "I\u0130\u0131i", textprofile.UnicodeDefaultCasefoldV1},
		{"accent", api.TextCasefoldUnicodeDefaultV1, api.TextAccentInsensitiveV1,
			api.OperatorStartsWith, "CAF\u00c9", textprofile.AccentInsensitiveV1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			fieldType := mustTextType(t, test.casefold, test.accent)
			values, err := ResolveLiterals(
				fieldType, test.operator,
				[]api.Literal{mustTextLiteral(t, test.input)},
			)
			if err != nil {
				t.Fatal(err)
			}
			requireScalars(t, values, test.normalize(test.input))
		})
	}
}
