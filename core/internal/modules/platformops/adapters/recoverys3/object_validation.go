package recoverys3

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var objectDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func validBucket(value string) bool {
	return len(value) >= 3 && len(value) <= 63 &&
		value == strings.ToLower(value) &&
		!strings.ContainsAny(value, "\x00\r\n/\\")
}

func validObjectKey(value string) bool {
	return len(value) >= 1 && len(value) <= 1024 &&
		value[0] != '/' && !strings.ContainsAny(value, "\x00\r\n\\") &&
		!strings.Contains(value, "//") &&
		!strings.Contains(value, "/../") &&
		!strings.HasPrefix(value, "../") &&
		!strings.HasSuffix(value, "/..")
}

func validVersion(value string) bool {
	return len(value) >= 1 && len(value) <= 1024 &&
		!strings.ContainsAny(value, "\x00\r\n")
}

func validExpected(digest string, bytes int64) bool {
	return objectDigestPattern.MatchString(digest) && bytes > 0
}

func validOperationID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}
