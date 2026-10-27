package serviceprincipal

import "regexp"

var capabilityKeyPattern = regexp.MustCompile(
	`^[a-z][a-z0-9]*([._-][a-z0-9]+)*$`,
)

type CapabilityKey string

func ParseCapabilityKey(raw string) (CapabilityKey, error) {
	if len(raw) < 3 || len(raw) > 100 || !capabilityKeyPattern.MatchString(raw) {
		return "", ErrInvalidGrant
	}
	return CapabilityKey(raw), nil
}
