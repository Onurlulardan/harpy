package projections

import (
	"encoding/hex"
	"sort"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func validProjectionInput(item ports.ProjectionItem, requirePermission bool) bool {
	if !item.ID.Valid() || !item.SourceID.Valid() || !item.SourceVersionID.Valid() ||
		item.SourceOwner == domain.SourceWork || !item.SourceOwner.Valid() ||
		item.SourceKind == "" || strings.TrimSpace(item.SourceKind) != item.SourceKind ||
		len(item.SourceKind) > 64 || len(item.Summary) > 500 ||
		item.SourceRevision < 1 || item.ReferenceEpoch < 1 || item.EventSequence < 1 ||
		(requirePermission && item.PermissionRevision < 1) ||
		!validProjectionPrincipals(item.VisibleTo) {
		return false
	}
	if item.DueAt != nil && item.DueAt.Location() != time.UTC {
		return false
	}
	if item.Surface == "INBOX" {
		return item.Window == nil
	}
	return item.Surface == "CALENDAR" && item.DueAt == nil && item.Window != nil &&
		item.Window.Validate() == nil
}

func projectionFingerprint(item ports.ProjectionItem, includePermission bool) [32]byte {
	parts := []string{"work-projection-item.v1", item.ID.String(),
		string(item.SourceOwner), item.SourceKind, item.SourceID.String(),
		item.SourceVersionID.String(), ports.VersionString(item.SourceRevision),
		ports.VersionString(item.ReferenceEpoch), ports.VersionString(item.EventSequence),
		item.Surface, item.Summary, projectionTime(item.DueAt), projectionWindow(item.Window)}
	if includePermission {
		parts = append(parts, ports.VersionString(item.PermissionRevision))
	}
	principals := make([]string, len(item.VisibleTo))
	for index, principal := range item.VisibleTo {
		principals[index] = string(principal.Kind) + ":" + principal.ID.String()
	}
	sort.Strings(principals)
	parts = append(parts, principals...)
	return ports.Fingerprint(parts...)
}

func repairFingerprint(owner domain.SourceOwner, kind string, sourceID domain.ID,
	through int64, items []ports.ProjectionItem,
) [32]byte {
	digests := make([]string, len(items))
	for index, item := range items {
		digest := projectionFingerprint(item, false)
		digests[index] = hex.EncodeToString(digest[:])
	}
	sort.Strings(digests)
	parts := []string{"work-projection-repair.v1", string(owner), kind,
		sourceID.String(), ports.VersionString(through), ports.VersionString(int64(len(items)))}
	return ports.Fingerprint(append(parts, digests...)...)
}

func projectionTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func projectionWindow(window *domain.CalendarWindow) string {
	if window == nil {
		return ""
	}
	return strings.Join([]string{projectionTime(window.StartsAt), projectionTime(window.EndsAt),
		window.StartDate, window.EndDate, window.TimeZone}, "|")
}
