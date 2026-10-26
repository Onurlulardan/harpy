package resolved

import (
	"strconv"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/textprofile"
)

func valueConformsToFieldType(
	fieldType api.FieldType, operator api.Operator, value Value,
) bool {
	kind, err := fieldType.Kind()
	if err != nil {
		return false
	}
	actual, err := value.Kind()
	if err != nil || actual != predicateValueKind(kind) {
		return false
	}
	scalar, scalarFound, scalarErr := value.Scalar()
	if scalarErr != nil {
		return false
	}
	switch kind {
	case api.LogicalTypeInteger:
		return scalarFound && integerWithinProfile(fieldType, scalar)
	case api.LogicalTypeDecimal:
		return scalarFound && decimalWithinProfile(fieldType, scalar)
	case api.LogicalTypeEnum:
		return scalarFound && enumWithinProfile(fieldType, scalar)
	case api.LogicalTypeText:
		return scalarFound && textMatchesOperatorProfile(fieldType, operator, scalar)
	default:
		return true
	}
}

func integerWithinProfile(fieldType api.FieldType, scalar string) bool {
	parsed, err := strconv.ParseInt(scalar, 10, 64)
	if err != nil {
		return false
	}
	profile, found, err := fieldType.Integer()
	if err != nil || !found {
		return false
	}
	minimum, minimumErr := profile.Minimum()
	maximum, maximumErr := profile.Maximum()
	return minimumErr == nil && maximumErr == nil &&
		parsed >= minimum && parsed <= maximum
}

func decimalWithinProfile(fieldType api.FieldType, scalar string) bool {
	profile, found, err := fieldType.Decimal()
	if err != nil || !found {
		return false
	}
	precision, precisionErr := profile.Precision()
	scale, scaleErr := profile.Scale()
	if precisionErr != nil || scaleErr != nil {
		return false
	}
	unsigned := strings.TrimPrefix(scalar, "-")
	parts := strings.Split(unsigned, ".")
	if scale == 0 && len(parts) != 1 || scale > 0 &&
		(len(parts) != 2 || len(parts[1]) != scale) {
		return false
	}
	wholeDigits := len(parts[0])
	if parts[0] == "0" {
		wholeDigits = 0
	}
	return wholeDigits <= precision-scale
}

func enumWithinProfile(fieldType api.FieldType, scalar string) bool {
	profile, found, err := fieldType.Enum()
	if err != nil || !found {
		return false
	}
	_, found, err = profile.Lookup(scalar)
	return err == nil && found
}

func textMatchesOperatorProfile(
	fieldType api.FieldType, operator api.Operator, scalar string,
) bool {
	if operator != api.OperatorContains && operator != api.OperatorStartsWith &&
		operator != api.OperatorEndsWith {
		return textprofile.ExactNFC(scalar) == scalar
	}
	profile, found, err := fieldType.Text()
	if err != nil || !found {
		return false
	}
	casefold, enabled, err := profile.Casefold()
	if err != nil || !enabled || casefold != api.TextCasefoldUnicodeDefaultV1 {
		return false
	}
	normalized := textprofile.UnicodeDefaultCasefoldV1(scalar)
	accent, hasAccent, err := profile.Accent()
	if err != nil {
		return false
	}
	if hasAccent {
		if accent != api.TextAccentInsensitiveV1 {
			return false
		}
		normalized = textprofile.AccentInsensitiveV1(scalar)
	}
	return normalized != "" && normalized == scalar
}
