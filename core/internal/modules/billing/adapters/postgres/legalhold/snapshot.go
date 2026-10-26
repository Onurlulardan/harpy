package legalhold

import (
	"fmt"
	"time"

	legalholdapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/legalhold"
	"github.com/google/uuid"
)

func scanSnapshot(row Row) (legalholdapp.Snapshot, error) {
	var result legalholdapp.Snapshot
	var releasedBy nullableUUID
	var releaseOperation nullableUUID
	var releaseReason nullableText
	var releasedAt nullableTime
	err := row.Scan(
		&result.HoldID, &result.TenantID, &result.State, &result.PlacedBy,
		&result.PlaceOperationID, &result.Reason, &result.Ticket, &result.PlacedAt,
		&releasedBy, &releaseOperation, &releaseReason, &releasedAt,
		&result.Revision, &result.Replay,
	)
	if releasedBy.Valid {
		result.ReleasedBy = releasedBy.value
	}
	if releaseOperation.Valid {
		result.ReleaseOperationID = releaseOperation.value
	}
	if releaseReason.Valid {
		result.ReleaseReason = releaseReason.value
	}
	if releasedAt.Valid {
		result.ReleasedAt = releasedAt.value.UTC()
	}
	result.PlacedAt = result.PlacedAt.UTC()
	return result, err
}

type nullableUUID struct {
	value uuid.UUID
	Valid bool
}

func (value *nullableUUID) Scan(source any) error {
	if source == nil {
		value.value, value.Valid = uuid.Nil, false
		return nil
	}
	var raw string
	switch selected := source.(type) {
	case uuid.UUID:
		value.value, value.Valid = selected, true
		return nil
	case string:
		raw = selected
	case []byte:
		raw = string(selected)
	default:
		return fmt.Errorf("billing legal hold nullable uuid is invalid")
	}
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return err
	}
	value.value, value.Valid = parsed, true
	return nil
}

type nullableText struct {
	value string
	Valid bool
}

func (value *nullableText) Scan(source any) error {
	if source == nil {
		value.value, value.Valid = "", false
		return nil
	}
	switch selected := source.(type) {
	case string:
		value.value = selected
	case []byte:
		value.value = string(selected)
	default:
		return fmt.Errorf("billing legal hold nullable text is invalid")
	}
	value.Valid = true
	return nil
}

type nullableTime struct {
	value time.Time
	Valid bool
}

func (value *nullableTime) Scan(source any) error {
	if source == nil {
		value.value, value.Valid = time.Time{}, false
		return nil
	}
	parsed, ok := source.(time.Time)
	if !ok {
		return fmt.Errorf("billing legal hold nullable time is invalid")
	}
	value.value, value.Valid = parsed, true
	return nil
}
