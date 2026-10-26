package v1adoption

import (
	"context"
	"time"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

func (source *LegacySource) readMemberships(ctx context.Context) ([]adoption.SourceRecord, error) {
	rows, err := queryRows(ctx, source, readMembershipsSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []adoption.SourceRecord{}
	for rows.Next() {
		var id, boardID, userID uuid.UUID
		var role string
		var created, updated *time.Time
		if err = rows.Scan(&id, &boardID, &userID, &role, &created, &updated); err != nil {
			return nil, ports.ErrUnavailable
		}
		value := adoption.LegacyMembership{ID: domain.ID(id), BoardID: domain.ID(boardID),
			UserID: domain.ID(userID), Role: role, CreatedAt: timeValue(created),
			UpdatedAt: timeValue(updated)}
		record, recordErr := sourceRecord(adoption.KindMembership, value.ID, value)
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
