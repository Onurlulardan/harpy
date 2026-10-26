package legalhold

import (
	"time"

	"github.com/google/uuid"
)

type rowStub struct {
	values []any
	err    error
}

func (row rowStub) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	for index, value := range row.values {
		switch destination := destinations[index].(type) {
		case *uuid.UUID:
			*destination = value.(uuid.UUID)
		case *string:
			*destination = value.(string)
		case *time.Time:
			*destination = value.(time.Time)
		case *nullableUUID:
			*destination = value.(nullableUUID)
		case *nullableText:
			*destination = value.(nullableText)
		case *nullableTime:
			*destination = value.(nullableTime)
		case *int64:
			*destination = value.(int64)
		case *bool:
			*destination = value.(bool)
		}
	}
	return nil
}

var adapterNow = time.Date(2026, 7, 17, 11, 0, 0, 0, time.UTC)

func activeRow(
	holdID, tenantID, actorID, operationID uuid.UUID,
) []any {
	return []any{
		holdID, tenantID, "active", actorID, operationID,
		"litigation preservation", "LEGAL-2026-41", adapterNow,
		nullableUUID{}, nullableUUID{}, nullableText{}, nullableTime{},
		int64(1), false,
	}
}
