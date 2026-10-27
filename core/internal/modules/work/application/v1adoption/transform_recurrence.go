package v1adoption

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) transformTaskRecurrence(ctx context.Context, scope ports.Scope,
	plan *Plan, source SourceRecord, board LegacyBoard, task LegacyTask,
) error {
	if !task.IsRecurring && task.RecurrenceRule == "" && task.RecurrenceEnd == nil {
		return nil
	}
	if !task.IsRecurring || task.DueDate == nil || task.RecurrenceEnd == nil ||
		!task.RecurrenceEnd.After(*task.DueDate) {
		appendIssue(plan, source, IssueRecurrence)
		return nil
	}
	location, err := service.dependencies.TZDB.Load(ctx, "UTC", service.dependencies.TZDBVersion)
	dtstart := task.DueDate.UTC().Format("20060102T150405")
	if err == nil {
		_, err = domain.ParseRecurrence(domain.RecurrenceInput{DTStart: dtstart,
			TimeZone: "UTC", TZDBVersion: service.dependencies.TZDBVersion,
			RRULE: task.RecurrenceRule, Location: location})
	}
	if err != nil {
		appendIssue(plan, source, IssueRecurrence)
		return nil
	}
	eventID := derivedID("recurring-task-event", task.ID.String())
	startsAt := task.DueDate.UTC()
	endsAt := startsAt.Add(time.Hour)
	event := TargetCalendarDocument{OwnerID: task.CreatedBy.String(), Audience: "BOARD",
		BoardID: task.BoardID.String(), Title: task.Title, Description: task.Description,
		Color: "#64748b", Timezone: "UTC", StartsAt: &startsAt, EndsAt: &endsAt,
		CreatedAt: utcPointer(&task.CreatedAt), UpdatedAt: utcPointer(&task.UpdatedAt),
		DeletedAt: utcPointer(task.DeletedAt)}
	if err = appendTarget(plan, source, TargetCalendar, eventID, event); err != nil {
		return err
	}
	ruleID := derivedID("recurring-task-rule", task.ID.String())
	rule := TargetRecurrenceDocument{EventID: eventID.String(), DTStart: dtstart,
		Timezone: "UTC", TZDBVersion: service.dependencies.TZDBVersion,
		RRULE: task.RecurrenceRule, HorizonEnd: task.RecurrenceEnd.UTC()}
	_ = board
	return appendTarget(plan, source, TargetRecurrence, ruleID, rule)
}
