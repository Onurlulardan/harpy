package billing

import "regexp"

var platformLicenseIdempotencyPattern = regexp.MustCompile(
	`^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$`,
)

func platformLicenseIdempotencyKey(values []string) (string, bool) {
	if len(values) != 1 ||
		!platformLicenseIdempotencyPattern.MatchString(values[0]) {
		return "", false
	}
	return values[0], true
}
