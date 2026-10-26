package v1adoption

import (
	"context"
	"time"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

func (source *LegacySource) readBoards(ctx context.Context) ([]adoption.SourceRecord, error) {
	rows, err := queryRows(ctx, source, readBoardsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []adoption.SourceRecord{}
	for rows.Next() {
		var id, organizationID, creatorID uuid.UUID
		var name, description string
		var created, updated, deleted *time.Time
		if err = rows.Scan(&id, &organizationID, &creatorID, &name, &description,
			&created, &updated, &deleted); err != nil {
			return nil, ports.ErrUnavailable
		}
		value := adoption.LegacyBoard{ID: domain.ID(id), OrganizationID: domain.ID(organizationID),
			CreatedBy: domain.ID(creatorID), Name: name, Description: description,
			CreatedAt: timeValue(created), UpdatedAt: timeValue(updated), DeletedAt: utcTime(deleted)}
		record, recordErr := sourceRecord(adoption.KindBoard, value.ID, value)
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

func (source *LegacySource) readColumns(ctx context.Context) ([]adoption.SourceRecord, error) {
	rows, err := queryRows(ctx, source, readColumnsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []adoption.SourceRecord{}
	for rows.Next() {
		var id, boardID uuid.UUID
		var name, color, status string
		var order int64
		var created, updated *time.Time
		if err = rows.Scan(&id, &boardID, &name, &color, &status, &order, &created, &updated); err != nil {
			return nil, ports.ErrUnavailable
		}
		value := adoption.LegacyColumn{ID: domain.ID(id), BoardID: domain.ID(boardID),
			Name: name, Color: color, LinkedStatus: status, Order: order,
			CreatedAt: timeValue(created), UpdatedAt: timeValue(updated)}
		record, recordErr := sourceRecord(adoption.KindColumn, value.ID, value)
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

func timeValue(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return value.UTC()
}

func utcTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	result := value.UTC()
	return &result
}
