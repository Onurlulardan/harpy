package v1adoption

import (
	"context"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) transformCalendar(ctx context.Context, scope ports.Scope,
	plan *Plan, index legacyIndex,
) error {
	for id, event := range index.calendar {
		source := index.source(KindCalendar, id)
		color := event.Color
		if color == "" {
			color = domain.DefaultColor
		}
		if !event.CreatedBy.Valid() || !event.OrganizationID.Valid() ||
			strings.TrimSpace(event.Title) == "" || len(event.Title) > 300 ||
			len(event.Description) > 16_000 || !event.EndDate.After(event.StartDate) ||
			!domain.ValidColor(color) || event.CreatedAt.IsZero() || event.UpdatedAt.IsZero() {
			appendIssue(plan, source, IssueFutureWork)
			continue
		}
		document := TargetCalendarDocument{OwnerID: event.CreatedBy.String(),
			Audience: "ORGANIZATION", OrganizationID: event.OrganizationID.String(),
			Title: event.Title, Description: event.Description, Color: color,
			Timezone:  "UTC",
			DeletedAt: utcPointer(event.DeletedAt), CreatedAt: utcPointer(&event.CreatedAt),
			UpdatedAt: utcPointer(&event.UpdatedAt)}
		if event.AllDay {
			document.StartDate = event.StartDate.UTC().Format("2006-01-02")
			document.EndDate = event.EndDate.UTC().Format("2006-01-02")
		} else {
			document.StartsAt, document.EndsAt = utcPointer(&event.StartDate), utcPointer(&event.EndDate)
		}
		if err := appendTarget(plan, source, TargetCalendar, id, document); err != nil {
			return err
		}
		if err := service.transformCalendarForm(ctx, scope, plan, source, event); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) transformCalendarForm(ctx context.Context, scope ports.Scope,
	plan *Plan, source SourceRecord, event LegacyCalendar,
) error {
	if event.FormSlug == "" && event.FormDataID == "" {
		return nil
	}
	var dataID *domain.ID
	if event.FormDataID != "" {
		parsed, err := domain.ParseID(event.FormDataID)
		if err != nil {
			appendIssue(plan, source, IssueUnresolvedForm)
			return nil
		}
		dataID = &parsed
	}
	reference, err := service.dependencies.Forms.ResolveForm(ctx, scope,
		event.FormSlug, dataID, nil)
	if err != nil || !reference.ID.Valid() || !reference.VersionID.Valid() ||
		reference.PermissionRevision < 1 {
		appendIssue(plan, source, IssueUnresolvedForm)
		return nil
	}
	document := TargetFormLinkDocument{ResourceKind: "CALENDAR_EVENT",
		ResourceID: event.ID.String(), FormID: reference.ID.String(),
		VersionID:          reference.VersionID.String(),
		PermissionRevision: reference.PermissionRevision}
	id := derivedID("calendar-form-link", event.ID.String(), reference.ID.String())
	return appendTarget(plan, source, TargetFormLink, id, document)
}
