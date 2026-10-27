package v1adoption

import (
	"encoding/json"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1adoption"
)

type sourceScanner interface{ Scan(...any) error }

func scanSourceRecord(row sourceScanner) (app.SourceRecord, error) {
	var value app.SourceRecord
	var action, entity *string
	var data []byte
	err := row.Scan(&value.ID, &value.TenantID, &value.UserID,
		&value.NotificationType, &value.Level, &value.Title, &value.Body,
		&action, &value.EntityID, &entity, &data, &value.Channel, &value.Read,
		&value.CreatedAt, &value.ReadAt, &value.InboxRevision)
	if err != nil {
		return app.SourceRecord{}, err
	}
	if action != nil {
		value.Action = *action
	}
	if entity != nil {
		value.Entity = *entity
	}
	if len(data) != 0 {
		value.Data = json.RawMessage(append([]byte(nil), data...))
	}
	return value, nil
}
