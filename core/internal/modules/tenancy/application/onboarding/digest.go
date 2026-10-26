package onboarding

import "regexp"

var digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func validDigest(value string) bool { return digestPattern.MatchString(value) }
