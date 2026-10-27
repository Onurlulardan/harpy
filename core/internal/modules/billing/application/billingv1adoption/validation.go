package billingv1adoptionapp

import (
	"encoding/hex"
	"regexp"
	"time"
)

var monthPattern = regexp.MustCompile(`^(\d{4})-(0[1-9]|1[0-2])$`)

func canonicalInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%1_000 == 0
}

func canonicalOptional(value *time.Time) bool {
	return value == nil || canonicalInstant(*value)
}

func validDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256Bytes && value == hex.EncodeToString(decoded)
}

func monthPeriod(value string) (time.Time, time.Time, bool) {
	if !monthPattern.MatchString(value) {
		return time.Time{}, time.Time{}, false
	}
	start, err := time.ParseInLocation("2006-01", value, time.UTC)
	if err != nil {
		return time.Time{}, time.Time{}, false
	}
	return start, start.AddDate(0, 1, 0), true
}

const sha256Bytes = 32
