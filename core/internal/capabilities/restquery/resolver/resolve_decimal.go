package resolver

import (
	"strings"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveDecimal(
	fieldType api.FieldType, source literalSource,
) (resolved.Value, restqueryv1.IssueCode) {
	text, issue := source.stringValue()
	if issue != "" {
		return resolved.Value{}, issue
	}
	negative, whole, fraction, valid := decimalParts(text)
	if !valid {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	profile, found, err := fieldType.Decimal()
	if err != nil || !found {
		return resolved.Value{}, restqueryv1.IssueTypeMismatch
	}
	precision, precisionErr := profile.Precision()
	scale, scaleErr := profile.Scale()
	if precisionErr != nil || scaleErr != nil {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	wholeDigits := len(whole)
	if whole == "0" {
		wholeDigits = 0
	}
	if wholeDigits > precision-scale || len(fraction) > scale {
		return resolved.Value{}, restqueryv1.IssueValueOutOfRange
	}
	canonical := whole
	if scale > 0 {
		canonical += "." + fraction + strings.Repeat("0", scale-len(fraction))
	}
	if negative {
		canonical = "-" + canonical
	}
	value, err := resolved.NewScalarValue(api.LogicalTypeDecimal, canonical)
	if err != nil {
		return resolved.Value{}, restqueryv1.IssueValueNotCanonical
	}
	return value, ""
}

func decimalParts(value string) (bool, string, string, bool) {
	if value == "" || value[0] == '+' || strings.ContainsAny(value, "eE") {
		return false, "", "", false
	}
	negative := value[0] == '-'
	unsigned := strings.TrimPrefix(value, "-")
	parts := strings.Split(unsigned, ".")
	if len(parts) > 2 || parts[0] == "" || len(parts) == 2 && parts[1] == "" ||
		len(parts[0]) > 1 && parts[0][0] == '0' {
		return false, "", "", false
	}
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	for _, part := range parts {
		for _, digit := range part {
			if digit < '0' || digit > '9' {
				return false, "", "", false
			}
		}
	}
	if negative && !strings.ContainsAny(unsigned, "123456789") {
		return false, "", "", false
	}
	return negative, parts[0], fraction, true
}
