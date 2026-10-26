package application

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

const (
	RetentionJobKind         = "usage.retention.prune"
	RetentionPayloadRevision = 1
	RetentionBatchLimit      = 1000
	RetentionAge             = 400 * 24 * time.Hour
	RetentionCycleInterval   = 24 * time.Hour
)

func RetentionReference(
	environmentID uuid.UUID,
	cycleAt time.Time,
) (string, string, error) {
	if environmentID == uuid.Nil || !validRetentionTime(cycleAt) {
		return "", "", domain.ErrInvalid
	}
	reference := "usage-retention/v1/" + environmentID.String() + "/" +
		strconv.FormatInt(cycleAt.UnixMicro(), 10)
	digest := sha256.Sum256([]byte(reference))
	return reference, hex.EncodeToString(digest[:]), nil
}

func ParseRetentionReference(reference string) (uuid.UUID, time.Time, error) {
	parts := strings.Split(reference, "/")
	if len(parts) != 4 || parts[0] != "usage-retention" || parts[1] != "v1" {
		return uuid.Nil, time.Time{}, domain.ErrInvalid
	}
	environmentID, err := uuid.Parse(parts[2])
	if err != nil || environmentID == uuid.Nil || parts[2] != environmentID.String() {
		return uuid.Nil, time.Time{}, domain.ErrInvalid
	}
	microseconds, err := strconv.ParseInt(parts[3], 10, 64)
	if err != nil || strconv.FormatInt(microseconds, 10) != parts[3] {
		return uuid.Nil, time.Time{}, domain.ErrInvalid
	}
	cycleAt := time.UnixMicro(microseconds).UTC()
	if !validRetentionTime(cycleAt) {
		return uuid.Nil, time.Time{}, domain.ErrInvalid
	}
	return environmentID, cycleAt, nil
}

func validRetentionTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%int(time.Microsecond) == 0
}
