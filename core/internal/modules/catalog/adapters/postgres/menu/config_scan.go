package menu

import (
	"encoding/json"
	"time"

	menuapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/menu"
	"github.com/google/uuid"
)

func (storage *store) scanConfig(source row) (menuapp.ConfigDocument, error) {
	var idRaw, createdByRaw, updatedByRaw string
	var name string
	var config []byte
	var active bool
	var createdAt, updatedAt time.Time
	if err := source.Scan(&idRaw, &name, &config, &active, &createdByRaw, &updatedByRaw,
		&createdAt, &updatedAt); err != nil {
		if storage.notFound(err) {
			return menuapp.ConfigDocument{}, menuapp.ErrNotFound
		}
		return menuapp.ConfigDocument{}, unavailable()
	}
	id, err := uuid.Parse(idRaw)
	createdBy, createdErr := uuid.Parse(createdByRaw)
	updatedBy, updatedErr := uuid.Parse(updatedByRaw)
	if err != nil || createdErr != nil || updatedErr != nil || id == uuid.Nil ||
		createdBy == uuid.Nil || updatedBy == uuid.Nil || !json.Valid(config) {
		return menuapp.ConfigDocument{}, unavailable()
	}
	return menuapp.ConfigDocument{
		ID: id, Name: name, Config: json.RawMessage(config), Active: active,
		CreatedBy: createdBy, UpdatedBy: updatedBy, CreatedAt: createdAt, UpdatedAt: updatedAt,
	}, nil
}

func validConfigWrite(value menuapp.ConfigMutation) bool {
	return value.ID != uuid.Nil && value.ID.Version() == 7 &&
		len(value.Name) >= 1 && len(value.Name) <= 128 && json.Valid(value.Config) &&
		value.ActorID != uuid.Nil && canonicalTime(value.OccurredAt)
}
