package cursor

import "time"

const (
	cursorVersion                = "v1"
	cursorSegments               = 4
	nonceBytes                   = 12
	maximumPlaintextBytes        = 8 * 1024
	maximumCursorBytes           = 16 * 1024
	maximumMandatoryBytes        = 64 * 1024
	maximumSorts                 = 6
	maximumCanonicalValueBytes   = 2048
	maximumPageSize              = 500
	cursorLifetime               = 15 * time.Minute
	allowedClockSkew             = 60 * time.Second
	mandatoryPredicateDigestInfo = "harpy/rest-query/cursor-mandatory-predicate/v1"
	redacted                     = "[REDACTED]"
)
