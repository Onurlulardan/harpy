package v1adoption

import (
	"context"
	"time"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (writer *CanonicalWriter) persistCollaborationRecord(ctx context.Context,
	record adoption.TargetRecord,
) error {
	switch record.Kind {
	case adoption.TargetComment:
		return writer.persistComment(ctx, record)
	case adoption.TargetCalendar:
		return writer.persistCalendar(ctx, record)
	case adoption.TargetRecurrence:
		return writer.persistRecurrence(ctx, record)
	case adoption.TargetProjection:
		return writer.persistProjection(ctx, record)
	default:
		return ports.ErrInvalid
	}
}

func (writer *CanonicalWriter) persistComment(ctx context.Context,
	record adoption.TargetRecord,
) error {
	var value adoption.TargetCommentDocument
	if err := decodeDocument(record, &value); err != nil {
		return err
	}
	boardID, err := parseID(value.BoardID)
	if err != nil {
		return err
	}
	taskID, err := parseID(value.TaskID)
	if err != nil {
		return err
	}
	authorID, err := parseID(value.AuthorID)
	if err != nil {
		return err
	}
	var redactedBy any
	if value.RedactedBy != "" {
		id, parseErr := parseID(value.RedactedBy)
		if parseErr != nil {
			return parseErr
		}
		redactedBy = id
	}
	if value.CreatedAt == nil {
		return ports.ErrInvalid
	}
	return writer.executeOne(ctx, adoptCommentSQL, record.ID, boardID, taskID, authorID,
		value.Body, value.State, nullable(value.RedactionKind), redactedBy,
		nullable(value.RedactionReason), *value.CreatedAt, value.RedactedAt)
}

func (writer *CanonicalWriter) persistCalendar(ctx context.Context,
	record adoption.TargetRecord,
) error {
	var value adoption.TargetCalendarDocument
	if err := decodeDocument(record, &value); err != nil ||
		value.CreatedAt == nil || value.UpdatedAt == nil {
		return ports.ErrInvalid
	}
	ownerID, err := parseID(value.OwnerID)
	if err != nil {
		return err
	}
	boardID, err := optionalID(value.BoardID)
	if err != nil {
		return err
	}
	organizationID, err := optionalID(value.OrganizationID)
	if err != nil {
		return err
	}
	return writer.executeOne(ctx, adoptCalendarSQL, record.ID, ownerID,
		value.Audience, boardID, organizationID, value.Title, value.Description,
		value.Color, value.StartsAt, value.EndsAt,
		nullable(value.StartDate), nullable(value.EndDate),
		value.Timezone, value.DeletedAt, *value.CreatedAt, *value.UpdatedAt)
}

func (writer *CanonicalWriter) persistRecurrence(ctx context.Context,
	record adoption.TargetRecord,
) error {
	var value adoption.TargetRecurrenceDocument
	if err := decodeDocument(record, &value); err != nil {
		return err
	}
	eventID, err := parseID(value.EventID)
	if err != nil {
		return err
	}
	effective, err := time.Parse("20060102T150405", value.DTStart)
	if err != nil {
		return ports.ErrInvalid
	}
	if err = writer.executeOne(ctx, adoptRecurrenceRuleSQL, record.ID, eventID); err != nil {
		return err
	}
	return writer.executeOne(ctx, adoptRecurrenceRevisionSQL, record.ID,
		value.DTStart, value.Timezone, value.TZDBVersion, value.RRULE, effective.UTC())
}

func (writer *CanonicalWriter) persistProjection(ctx context.Context,
	record adoption.TargetRecord,
) error {
	var value adoption.TargetProjectionDocument
	if err := decodeDocument(record, &value); err != nil {
		return err
	}
	sourceID, err := parseID(value.SourceID)
	if err != nil {
		return err
	}
	versionID, err := parseID(value.VersionID)
	if err != nil {
		return err
	}
	return writer.executeOne(ctx, adoptProjectionSQL, record.ID, sourceID, versionID,
		value.Revision, value.PermissionRevision, value.Summary, value.DueAt,
		value.VisibleToPrincipals)
}

func optionalID(value string) (any, error) {
	if value == "" {
		return nil, nil
	}
	return parseID(value)
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}
