package query

import (
	"crypto/sha256"
	"encoding/binary"
	"regexp"
	"time"

	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	"github.com/google/uuid"
)

var namePattern = regexp.MustCompile(`^[\pL\pN][\pL\pN _.-]{0,99}$`)
var idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,128}$`)

func validName(value string) bool { return namePattern.MatchString(value) }

func validIdempotency(value string) bool { return idempotencyPattern.MatchString(value) }

func validUUIDv7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 && value.Variant() == uuid.RFC4122
}

func validTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}

func requestDigest(parts ...[]byte) [32]byte {
	hash := sha256.New()
	for _, part := range parts {
		length := make([]byte, 8)
		binary.BigEndian.PutUint64(length, uint64(len(part)))
		_, _ = hash.Write(length)
		_, _ = hash.Write(part)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func validateSnapshot(value Snapshot) error {
	queryID, err := analyticsquery.NewID([16]byte(value.QueryID))
	if err != nil {
		return ErrUnavailable
	}
	versionID, err := analyticsquery.NewVersionID([16]byte(value.VersionID))
	if err != nil {
		return ErrUnavailable
	}
	if _, err = analyticsquery.Rehydrate(queryID, value.Lifecycle, value.RootRevision,
		value.ReferenceEpoch, versionID); err != nil || !validName(value.Name) ||
		!value.CachePolicy.Valid() ||
		!validTime(value.CreatedAt) || !validTime(value.UpdatedAt) ||
		value.UpdatedAt.Before(value.CreatedAt) {
		return ErrUnavailable
	}
	queryDocument, queryErr := NewDocument(value.QueryDocument)
	_, parameterErr := NewDocument(value.ParameterDocument)
	_, outputErr := NewDocument(value.OutputDocument)
	if queryErr != nil || parameterErr != nil || outputErr != nil {
		return ErrUnavailable
	}
	_, err = analyticsquery.RehydrateVersion(versionID, queryID, value.VersionNumber,
		value.VersionRevision, value.VersionState, queryDocument.SHA256(), value.ManifestSHA256)
	return mapDomainError(err)
}

func mapDomainError(err error) error {
	if err == nil {
		return nil
	}
	if err == analyticsquery.ErrConflict {
		return ErrConflict
	}
	return ErrUnavailable
}
