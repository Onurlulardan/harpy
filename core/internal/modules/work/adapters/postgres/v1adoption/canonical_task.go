package v1adoption

import (
	"context"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (writer *CanonicalWriter) persistTaskRecord(ctx context.Context,
	record adoption.TargetRecord,
) error {
	switch record.Kind {
	case adoption.TargetTask:
		return writer.persistTask(ctx, record)
	case adoption.TargetAssignee:
		var value adoption.TargetAssigneeDocument
		if err := decodeDocument(record, &value); err != nil {
			return err
		}
		taskID, err := parseID(value.TaskID)
		if err != nil {
			return err
		}
		principalID, err := parseID(value.PrincipalID)
		if err != nil {
			return err
		}
		return writer.executeOne(ctx, adoptAssigneeSQL, taskID, value.PrincipalKind, principalID)
	case adoption.TargetLabel:
		var value adoption.TargetLabelDocument
		if err := decodeDocument(record, &value); err != nil {
			return err
		}
		boardID, err := parseID(value.BoardID)
		if err != nil {
			return err
		}
		return writer.executeOne(ctx, adoptLabelSQL, record.ID, boardID, value.Name, value.Color)
	case adoption.TargetTaskLabel:
		var value adoption.TargetTaskLabelDocument
		if err := decodeDocument(record, &value); err != nil {
			return err
		}
		taskID, err := parseID(value.TaskID)
		if err != nil {
			return err
		}
		labelID, err := parseID(value.LabelID)
		if err != nil {
			return err
		}
		return writer.executeOne(ctx, adoptTaskLabelSQL, taskID, labelID)
	case adoption.TargetChecklist:
		var value adoption.TargetChecklistDocument
		if err := decodeDocument(record, &value); err != nil {
			return err
		}
		taskID, err := parseID(value.TaskID)
		if err != nil {
			return err
		}
		return writer.executeOne(ctx, adoptChecklistSQL, record.ID, taskID,
			value.Text, value.Completed, value.Position)
	case adoption.TargetFormLink:
		return writer.persistFormLink(ctx, record)
	default:
		return ports.ErrInvalid
	}
}

func (writer *CanonicalWriter) persistTask(ctx context.Context,
	record adoption.TargetRecord,
) error {
	var value adoption.TargetTaskDocument
	if err := decodeDocument(record, &value); err != nil {
		return err
	}
	boardID, err := parseID(value.BoardID)
	if err != nil {
		return err
	}
	columnID, err := parseID(value.ColumnID)
	if err != nil {
		return err
	}
	creatorID, err := parseID(value.CreatorID)
	if err != nil {
		return err
	}
	if value.CreatedAt == nil || value.UpdatedAt == nil {
		return ports.ErrInvalid
	}
	if err = writer.executeOne(ctx, adoptTaskSQL, record.ID, boardID, columnID, creatorID,
		value.Kind, value.Status, value.Priority, value.Title, value.Description, value.Position,
		value.StartsAt, value.DueAt, value.CompletedAt, value.CancelledAt, value.DeletedAt,
		*value.CreatedAt, *value.UpdatedAt); err != nil {
		return err
	}
	legacy, err := domain.NormalizeLegacyTaskMetadata(domain.LegacyTaskMetadata{
		IsRecurring: value.IsRecurring, RecurrenceRule: value.RecurrenceRule,
		RecurrenceEnd: value.RecurrenceEnd, FormFilter: value.FormFilter})
	if err != nil {
		return ports.ErrInvalid
	}
	return writer.executeOne(ctx, adoptTaskCompatibilityMetadataSQL, record.ID,
		legacy.IsRecurring, legacy.RecurrenceRule, legacy.RecurrenceEnd,
		[]byte(legacy.FormFilter), *value.CreatedAt, *value.UpdatedAt)
}

func (writer *CanonicalWriter) persistFormLink(ctx context.Context,
	record adoption.TargetRecord,
) error {
	var value adoption.TargetFormLinkDocument
	if err := decodeDocument(record, &value); err != nil {
		return err
	}
	resourceID, err := parseID(value.ResourceID)
	if err != nil {
		return err
	}
	formID, err := parseID(value.FormID)
	if err != nil {
		return err
	}
	versionID, err := parseID(value.VersionID)
	if err != nil {
		return err
	}
	return writer.executeOne(ctx, adoptFormLinkSQL, record.ID, value.ResourceKind,
		resourceID, formID, versionID, value.ReferenceEpoch, value.PermissionRevision)
}
