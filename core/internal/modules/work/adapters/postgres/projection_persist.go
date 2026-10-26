package postgres

import (
	"context"
	_ "embed"
	"sort"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

//go:embed sql/tenant/upsert_inbox_projection.sql
var upsertInboxProjectionSQL string

//go:embed sql/tenant/upsert_calendar_projection.sql
var upsertCalendarProjectionSQL string

func (store *Store) PersistProjection(ctx context.Context, item ports.ProjectionItem) error {
	if !validProjectionItem(item) {
		return application.ErrInvalid
	}
	state := "REVOKED"
	if item.Visible {
		state = "VISIBLE"
	}
	if item.Surface == "INBOX" {
		principals := projectionPrincipalKeys(item.VisibleTo)
		return executeOne(ctx, store, upsertInboxProjectionSQL, item.ID, item.SourceOwner,
			item.SourceKind, item.SourceID, item.SourceVersionID, item.SourceRevision,
			item.PermissionRevision, item.EventSequence, item.Summary, item.DueAt, state,
			principals)
	}
	window := *item.Window
	principals := projectionPrincipalKeys(item.VisibleTo)
	return executeOne(ctx, store, upsertCalendarProjectionSQL, item.ID, item.SourceOwner,
		item.SourceKind, item.SourceID, item.SourceVersionID, item.SourceRevision,
		item.PermissionRevision, item.EventSequence, item.Summary, window.StartsAt,
		window.EndsAt, nullText(window.StartDate), nullText(window.EndDate), state,
		principals)
}

func validProjectionItem(item ports.ProjectionItem) bool {
	if !item.ID.Valid() || !item.SourceID.Valid() || !item.SourceVersionID.Valid() ||
		item.SourceOwner == domain.SourceWork || !item.SourceOwner.Valid() ||
		strings.TrimSpace(item.SourceKind) == "" ||
		strings.TrimSpace(item.SourceKind) != item.SourceKind ||
		len(item.SourceKind) > 64 || len(item.Summary) > 500 || item.SourceRevision < 1 ||
		item.ReferenceEpoch < 1 ||
		item.PermissionRevision < 1 || item.EventSequence < 1 ||
		!validProjectionPrincipalSet(item.VisibleTo) ||
		(item.DueAt != nil && item.DueAt.Location() != time.UTC) {
		return false
	}
	if item.Surface == "INBOX" {
		return item.Window == nil
	}
	return item.Surface == "CALENDAR" && item.DueAt == nil && item.Window != nil &&
		item.Window.Validate() == nil
}

func validProjectionPrincipalSet(principals []ports.Principal) bool {
	if len(principals) == 0 || len(principals) > 100 {
		return false
	}
	seen := make(map[ports.Principal]struct{}, len(principals))
	for _, principal := range principals {
		if !principal.Valid() {
			return false
		}
		if _, duplicate := seen[principal]; duplicate {
			return false
		}
		seen[principal] = struct{}{}
	}
	return true
}

func projectionPrincipalKeys(principals []ports.Principal) []string {
	result := make([]string, len(principals))
	for index, principal := range principals {
		result[index] = string(principal.Kind) + ":" + principal.ID.String()
	}
	sort.Strings(result)
	return result
}
