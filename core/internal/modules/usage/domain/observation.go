package domain

import (
	"crypto/sha256"
	"encoding/binary"
	"strings"
	"time"
	"unicode/utf8"
)

type ID [16]byte

type Observation struct {
	ID, EnvironmentID, TenantID, ActorID ID
	SourceIdentity                       [32]byte
	Method, RoutePattern                 string
	StatusCode                           int
	DurationMicros                       int64
	OccurredAt                           time.Time
}

func NormalizeObservation(value Observation) (Observation, error) {
	value.Method = strings.ToUpper(strings.TrimSpace(value.Method))
	value.RoutePattern = strings.TrimSpace(value.RoutePattern)
	value.OccurredAt = value.OccurredAt.UTC()
	if !validV7(value.ID) || value.EnvironmentID == (ID{}) ||
		value.TenantID == (ID{}) || value.ActorID == (ID{}) ||
		zeroDigest(value.SourceIdentity) || !validMethod(value.Method) ||
		!validRoute(value.RoutePattern) || value.StatusCode < 100 ||
		value.StatusCode > 599 || value.DurationMicros < 0 ||
		value.DurationMicros > int64((24*time.Hour)/time.Microsecond) ||
		!storedTime(value.OccurredAt) {
		return Observation{}, ErrInvalid
	}
	return value, nil
}

func (value Observation) Digest() [32]byte {
	document := make([]byte, 0, 160)
	document = append(document, value.EnvironmentID[:]...)
	document = append(document, value.TenantID[:]...)
	document = append(document, value.ActorID[:]...)
	document = appendField(document, value.Method)
	document = appendField(document, value.RoutePattern)
	var numbers [24]byte
	binary.BigEndian.PutUint64(numbers[0:8], uint64(value.StatusCode))
	binary.BigEndian.PutUint64(numbers[8:16], uint64(value.DurationMicros))
	binary.BigEndian.PutUint64(numbers[16:24], uint64(value.OccurredAt.UnixMicro()))
	document = append(document, numbers[:]...)
	return sha256.Sum256(document)
}

func appendField(target []byte, value string) []byte {
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(value)))
	target = append(target, size[:]...)
	return append(target, value...)
}

func validV7(value ID) bool { return value != (ID{}) && value[6]>>4 == 7 }

func validMethod(value string) bool {
	switch value {
	case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS":
		return true
	default:
		return false
	}
}

func validRoute(value string) bool {
	return utf8.ValidString(value) && len(value) >= 1 && len([]byte(value)) <= 256 &&
		value[0] == '/' && !strings.ContainsAny(value, "?#\x00\r\n")
}

func ValidRoutePattern(value string) bool { return validRoute(value) }

func storedTime(value time.Time) bool {
	return !value.IsZero() && value.Nanosecond()%1000 == 0
}

func zeroDigest(value [32]byte) bool { return value == [32]byte{} }
