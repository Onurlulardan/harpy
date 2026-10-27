package v1adoption

import (
	"context"
	"encoding/json"
	"time"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

func (source *LegacySource) readTasks(ctx context.Context) ([]adoption.SourceRecord, error) {
	rows, err := queryRows(ctx, source, readTasksSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []adoption.SourceRecord{}
	for rows.Next() {
		value, scanErr := scanTask(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		record, recordErr := sourceRecord(adoption.KindTask, value.ID, value)
		if recordErr != nil {
			return nil, recordErr
		}
		result = append(result, record)
	}
	if rows.Err() != nil {
		return nil, ports.ErrUnavailable
	}
	return result, nil
}

func scanTask(rows Rows) (adoption.LegacyTask, error) {
	var id, boardID, columnID, creatorID uuid.UUID
	var formDataID, assigneeID, workflowTaskID *uuid.UUID
	var title, description, kind, priority, status, formSlug, assigneeKind, rule string
	var order int64
	var due, start, completed, recurrenceEnd, deleted, created, updated *time.Time
	var recurring bool
	var filter, labels, checklist []byte
	err := rows.Scan(&id, &boardID, &columnID, &creatorID, &title, &description,
		&kind, &priority, &status, &formSlug, &assigneeKind, &order,
		&formDataID, &assigneeID, &workflowTaskID, &due, &start, &completed,
		&recurrenceEnd, &deleted, &recurring, &rule, &filter, &labels, &checklist,
		&created, &updated)
	if err != nil {
		return adoption.LegacyTask{}, ports.ErrUnavailable
	}
	return adoption.LegacyTask{ID: domain.ID(id), BoardID: domain.ID(boardID),
		ColumnID: domain.ID(columnID), CreatedBy: domain.ID(creatorID), Title: title,
		Description: description, TaskType: kind, Priority: priority, Status: status,
		FormSlug: formSlug, AssigneeType: assigneeKind, ColumnOrder: order,
		FormDataID: domainID(formDataID), AssigneeID: domainID(assigneeID),
		WorkflowTaskID: domainID(workflowTaskID), DueDate: utcTime(due),
		StartDate: utcTime(start), CompletedAt: utcTime(completed),
		RecurrenceEnd: utcTime(recurrenceEnd), DeletedAt: utcTime(deleted),
		IsRecurring: recurring, RecurrenceRule: rule,
		FormFilter: cloneJSON(filter), Labels: cloneJSON(labels), Checklist: cloneJSON(checklist),
		CreatedAt: timeValue(created), UpdatedAt: timeValue(updated)}, nil
}

func domainID(value *uuid.UUID) *domain.ID {
	if value == nil {
		return nil
	}
	result := domain.ID(*value)
	return &result
}

func cloneJSON(value []byte) json.RawMessage { return append(json.RawMessage(nil), value...) }
