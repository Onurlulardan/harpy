package dashboardexport

import (
	"crypto/sha256"
	"encoding/binary"
	"regexp"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

var idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$`)

func canonicalParameters(raw []byte) ([]byte, error) {
	if len(raw) == 0 {
		raw = []byte(`{}`)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(raw, 64<<10)
	if err != nil {
		return nil, ErrInvalid
	}
	return canonical, nil
}

func validUUIDv7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 && value.Variant() == uuid.RFC4122
}

func validTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}

func requestDigest(operation string, tenantID, actorID uuid.UUID, parts ...[]byte) [32]byte {
	values := [][]byte{[]byte(operation), tenantID[:], actorID[:]}
	values = append(values, parts...)
	hash := sha256.New()
	for _, value := range values {
		length := make([]byte, 8)
		binary.BigEndian.PutUint64(length, uint64(len(value)))
		_, _ = hash.Write(length)
		_, _ = hash.Write(value)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func validateSnapshot(value Snapshot, dashboardID, exportID uuid.UUID) error {
	if value.ExportID != exportID || value.DashboardID != dashboardID ||
		!validUUIDv7(value.ExportID) || !validUUIDv7(value.DashboardID) ||
		!validUUIDv7(value.DashboardVersionID) || !validUUIDv7(value.JobID) ||
		!value.State.Valid() || !value.StoredState.Valid() || !value.Format.Valid() ||
		!validTime(value.CreatedAt) || !validTime(value.UpdatedAt) ||
		!validTime(value.ExpiresAt) || value.UpdatedAt.Before(value.CreatedAt) ||
		!value.ExpiresAt.After(value.CreatedAt) || value.RowCount < 0 ||
		value.PlaintextBytes < 0 {
		return ErrUnavailable
	}
	if value.State == StateReady && (value.ContentType != value.Format.ContentType() ||
		value.Filename == "" || value.PlaintextBytes < 1) {
		return ErrUnavailable
	}
	return nil
}
