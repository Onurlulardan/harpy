package api

import "regexp"

var onboardingDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func validOnboardingDigest(value string) bool {
	return onboardingDigestPattern.MatchString(value)
}
