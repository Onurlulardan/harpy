package application

import (
	"regexp"

	"github.com/google/uuid"
)

var recoveryDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func validRecoveryUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func validRecoveryDigest(value string) bool {
	return recoveryDigestPattern.MatchString(value)
}
