package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
)

type statsDocument struct {
	Total       int64            `json:"total"`
	Last24Hours int64            `json:"last_24_hours"`
	LastHour    int64            `json:"last_hour"`
	ByAction    map[string]int64 `json:"by_action"`
	ByEntity    map[string]int64 `json:"by_entity"`
	Actors      []struct {
		ActorID string `json:"actor_id"`
		Count   int64  `json:"count"`
	} `json:"actors"`
	Entities []struct {
		EntityType string `json:"entity_type"`
		EntityID   string `json:"entity_id"`
		FormSlug   string `json:"form_slug"`
		Count      int64  `json:"count"`
	} `json:"entities"`
}

func (store *Store) Stats(ctx context.Context, now time.Time) (domain.Stats, error) {
	if !store.allowed(ctx, application.ScopeSnapshot) || now.IsZero() ||
		now.Nanosecond()%1000 != 0 {
		return domain.Stats{}, application.ErrInvalid
	}
	var document []byte
	if err := store.queryRow(ctx, statsSQL, now).Scan(&document); err != nil {
		return domain.Stats{}, mapError(err, store.notFound)
	}
	var wire statsDocument
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&wire) != nil || wire.ByAction == nil || wire.ByEntity == nil {
		return domain.Stats{}, application.ErrUnavailable
	}
	stats := domain.Stats{Total: wire.Total, Last24Hours: wire.Last24Hours,
		LastHour: wire.LastHour, ByAction: wire.ByAction, ByEntity: wire.ByEntity,
		Actors:   make([]domain.ActorStat, 0, len(wire.Actors)),
		Entities: make([]domain.EntityStat, 0, len(wire.Entities))}
	for _, actor := range wire.Actors {
		id, err := domain.ParseID(actor.ActorID)
		if err != nil {
			return domain.Stats{}, application.ErrUnavailable
		}
		stats.Actors = append(stats.Actors, domain.ActorStat{ActorID: id, Count: actor.Count})
	}
	for _, entity := range wire.Entities {
		stats.Entities = append(stats.Entities, domain.EntityStat{EntityType: entity.EntityType,
			EntityID: entity.EntityID, FormSlug: entity.FormSlug, Count: entity.Count})
	}
	return stats, nil
}
