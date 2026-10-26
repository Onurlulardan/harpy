package resolved

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/textprofile"
)

func TestTextConformanceMatchesPlatformProfileAuthority(t *testing.T) {
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
			api.OperatorContains, "CAF\u00c9", textprofile.AccentInsensitiveV1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			profile, err := api.NewTextProfile(test.casefold, test.accent)
			if err != nil {
				t.Fatal(err)
			}
			fieldType, err := api.NewTextFieldType(profile)
			if err != nil {
				t.Fatal(err)
			}
			canonical := test.normalize(test.input)
			value, err := NewScalarValue(api.LogicalTypeText, canonical)
			if err != nil || !valueConformsToFieldType(fieldType, test.operator, value) {
				t.Fatalf("canonical platform value rejected: %v", err)
			}
			if test.input == canonical {
				return
			}
			raw, err := NewScalarValue(api.LogicalTypeText, test.input)
			if err == nil && valueConformsToFieldType(fieldType, test.operator, raw) {
				t.Fatal("non-canonical consumer value was accepted")
			}
		})
	}
}
