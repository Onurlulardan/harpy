package lifecycle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	"github.com/google/uuid"
)

var idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{7,127}$`)
var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func validIdempotency(value string) bool { return idempotencyPattern.MatchString(value) }

func requestDigest(value any) (string, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return "", ErrInvalid
	}
	document, err := canonicaljsonv1.CanonicalizeObject(raw, 16384)
	if err != nil {
		return "", ErrInvalid
	}
	digest := sha256.Sum256(document)
	return hex.EncodeToString(digest[:]), nil
}

func (application *Application) nextIdentity() (uuid.UUID, error) {
	value, err := application.ids.New()
	if err != nil || !validUUIDv7(value) {
		return uuid.Nil, ErrUnavailable
	}
	return value, nil
}

func stableIdentity(value uuid.UUID) (packagecontract.StableID, error) {
	identity, err := packagecontract.ParseStableID(value.String())
	if err != nil {
		return packagecontract.StableID{}, ErrInvalid
	}
	return identity, nil
}

func uuidFromStable(value packagecontract.StableID) (uuid.UUID, error) {
	if !value.Valid() {
		return uuid.Nil, ErrInvalid
	}
	parsed, err := uuid.Parse(value.String())
	if err != nil || parsed == uuid.Nil {
		return uuid.Nil, ErrInvalid
	}
	return parsed, nil
}

func uuidFromCanonical(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != value {
		return uuid.Nil, ErrInvalid
	}
	return parsed, nil
}

func (application *Application) now() (time.Time, error) {
	value := application.clock.Now()
	if value.IsZero() || value.Location() != time.UTC || value.Nanosecond()%1000 != 0 {
		return time.Time{}, ErrUnavailable
	}
	return value, nil
}
