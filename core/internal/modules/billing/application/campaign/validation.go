package campaign

import (
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

func validActor(value uuid.UUID) bool { return value != uuid.Nil }

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}

func validInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}

func validText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value
}

func validExpectedRevision(value int64) bool {
	return value > 0 && value < math.MaxInt64
}

func nextV7(nextID IDGenerator) (uuid.UUID, error) {
	value, err := nextID()
	if err != nil || !validV7(value) {
		return uuid.Nil, ErrConfiguration
	}
	return value, nil
}

func validCreateResult(value Snapshot, campaignID, versionID uuid.UUID) bool {
	if !validPublishedSnapshot(value) || value.Revision != 1 || value.Version != 1 {
		return false
	}
	return value.Replay || value.CampaignID == campaignID && value.VersionID == versionID
}

func validCreateVersionResult(
	value Snapshot,
	campaignID uuid.UUID,
	versionID uuid.UUID,
	expectedRevision int64,
) bool {
	if !validPublishedSnapshot(value) || value.CampaignID != campaignID ||
		value.Revision != expectedRevision+1 {
		return false
	}
	return value.Replay || value.VersionID == versionID
}

func validRetireResult(value Snapshot, campaignID uuid.UUID, expectedRevision int64) bool {
	return validV7(value.CampaignID) && value.CampaignID == campaignID &&
		value.VersionID == uuid.Nil && value.Revision == expectedRevision+1 &&
		value.Version == 0 && value.Lifecycle == "retired"
}

func validPublishedSnapshot(value Snapshot) bool {
	return validV7(value.CampaignID) && validV7(value.VersionID) &&
		value.Revision > 0 && value.Version > 0 && value.Lifecycle == "active"
}
