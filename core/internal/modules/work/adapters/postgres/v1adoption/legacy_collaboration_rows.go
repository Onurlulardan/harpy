package v1adoption

import (
	"context"
	"time"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

func (source *LegacySource) readComments(ctx context.Context) ([]adoption.SourceRecord, error) {
	rows, err := queryRows(ctx, source, readCommentsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []adoption.SourceRecord{}
	for rows.Next() {
		var id, taskID, userID uuid.UUID
		var content string
		var created, updated, deleted *time.Time
		if err = rows.Scan(&id, &taskID, &userID, &content, &created, &updated, &deleted); err != nil {
			return nil, ports.ErrUnavailable
		}
		value := adoption.LegacyComment{ID: domain.ID(id), TaskID: domain.ID(taskID),
			UserID: domain.ID(userID), Content: content, CreatedAt: timeValue(created),
			UpdatedAt: timeValue(updated), DeletedAt: utcTime(deleted)}
		record, recordErr := sourceRecord(adoption.KindComment, value.ID, value)
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

func (source *LegacySource) readCalendar(ctx context.Context) ([]adoption.SourceRecord, error) {
	rows, err := queryRows(ctx, source, readCalendarSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []adoption.SourceRecord{}
	for rows.Next() {
		var id, organizationID, creatorID uuid.UUID
		var title, description, color, formSlug, formDataID string
		var start, end, created, updated, deleted *time.Time
		var allDay bool
		if err = rows.Scan(&id, &organizationID, &creatorID, &title, &description,
			&color, &start, &end, &allDay, &formSlug, &formDataID,
			&created, &updated, &deleted); err != nil {
			return nil, ports.ErrUnavailable
		}
		value := adoption.LegacyCalendar{ID: domain.ID(id),
			OrganizationID: domain.ID(organizationID), CreatedBy: domain.ID(creatorID),
			Title: title, Description: description, Color: color,
			StartDate: timeValue(start), EndDate: timeValue(end), AllDay: allDay,
			FormSlug: formSlug, FormDataID: formDataID, CreatedAt: timeValue(created),
			UpdatedAt: timeValue(updated), DeletedAt: utcTime(deleted)}
		record, recordErr := sourceRecord(adoption.KindCalendar, value.ID, value)
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
