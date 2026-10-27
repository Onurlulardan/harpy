package audithttp

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
)

type fieldChangeView struct {
	Field, Classification       string
	Before, After               *string
	BeforePresent, AfterPresent bool
}

func (view fieldChangeView) MarshalJSON() ([]byte, error) {
	type wire struct {
		Field          string  `json:"field"`
		Classification string  `json:"classification"`
		Before         *string `json:"before,omitempty"`
		After          *string `json:"after,omitempty"`
		BeforePresent  bool    `json:"before_present"`
		AfterPresent   bool    `json:"after_present"`
	}
	return marshal(wire(view))
}

func entryView(entry domain.Entry) map[string]any {
	changes := make([]fieldChangeView, 0, len(entry.FieldChanges))
	for _, change := range entry.FieldChanges {
		changes = append(changes, fieldChangeView{Field: change.Field,
			Classification: string(change.Classification), Before: change.Before,
			After: change.After, BeforePresent: change.BeforeSHA256 != [32]byte{},
			AfterPresent: change.AfterSHA256 != [32]byte{}})
	}
	view := map[string]any{"id": entry.ID.String(), "owner_audit_id": entry.OwnerAuditID.String(),
		"owner_module": entry.OwnerModule, "operation": entry.Operation,
		"entity_type": entry.EntityType, "entity_id": entry.EntityID,
		"action": entry.Action, "user_id": entry.ActorID.String(),
		"authorization_version": entry.AuthorizationVersion,
		"field_changes":         changes, "created_at": wireTime(entry.OccurredAt),
		"recorded_at": wireTime(entry.RecordedAt)}
	if entry.OrganizationID != nil {
		view["org_id"] = entry.OrganizationID.String()
	}
	if entry.CorrelationID != nil {
		view["correlation_id"] = entry.CorrelationID.String()
	}
	if entry.FormSlug != "" {
		view["form_slug"] = entry.FormSlug
	}
	if entry.ReasonCode != "" {
		view["reason_code"] = entry.ReasonCode
	}
	return view
}

func pageView(page domain.Page) map[string]any {
	entries := make([]map[string]any, 0, len(page.Entries))
	for _, entry := range page.Entries {
		entries = append(entries, entryView(entry))
	}
	return map[string]any{"changes": entries, "total": page.Total, "page": page.Page,
		"limit": page.Limit, "total_pages": page.TotalPages}
}

func statsView(stats domain.Stats) map[string]any {
	actors := make([]map[string]any, 0, len(stats.Actors))
	for _, actor := range stats.Actors {
		actors = append(actors, map[string]any{"user_id": actor.ActorID.String(), "count": actor.Count})
	}
	entities := make([]map[string]any, 0, len(stats.Entities))
	for _, entity := range stats.Entities {
		item := map[string]any{"entity_type": entity.EntityType,
			"entity_id": entity.EntityID, "change_count": entity.Count}
		if entity.FormSlug != "" {
			item["form_slug"] = entity.FormSlug
		}
		entities = append(entities, item)
	}
	return map[string]any{"total_changes": stats.Total,
		"changes_by_action": stats.ByAction, "changes_by_entity": stats.ByEntity,
		"changes_by_user": actors, "changes_last_24h": stats.Last24Hours,
		"changes_last_hour": stats.LastHour, "most_active_entities": entities}
}

func wireTime(value time.Time) string {
	return value.UTC().Truncate(time.Microsecond).Format("2006-01-02T15:04:05.000000Z")
}
