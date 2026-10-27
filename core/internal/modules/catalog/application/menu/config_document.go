package menu

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maximumConfigItems = 2048

type ConfigDocument struct {
	ID                   uuid.UUID
	Name                 string
	Config               json.RawMessage
	Active               bool
	CreatedBy, UpdatedBy uuid.UUID
	CreatedAt, UpdatedAt time.Time
}

type ConfigMutation struct {
	ID         uuid.UUID
	Name       string
	Config     json.RawMessage
	Active     bool
	ActorID    uuid.UUID
	OccurredAt time.Time
}

type ConfigWriteRequest struct {
	ID             uuid.UUID
	Name           string
	Config         json.RawMessage
	Active         bool
	IdempotencyKey string
}

func normalizeConfigName(name string) (string, bool) {
	trimmed := strings.TrimSpace(name)
	return trimmed, len(trimmed) >= 1 && len(trimmed) <= 128
}

func validConfigJSON(raw json.RawMessage) bool {
	if !json.Valid(raw) {
		return false
	}
	var root map[string]json.RawMessage
	if json.Unmarshal(raw, &root) != nil {
		return false
	}
	if groups, found := root["groups"]; found && !jsonArray(groups) {
		return false
	}
	if ungrouped, found := root["ungrouped"]; found && !jsonArray(ungrouped) {
		return false
	}
	count := countConfigItems(raw)
	return count >= 0 && count <= maximumConfigItems
}

func jsonArray(raw json.RawMessage) bool {
	var value []json.RawMessage
	return json.Unmarshal(raw, &value) == nil
}

func countConfigItems(raw json.RawMessage) int {
	var root map[string]any
	if json.Unmarshal(raw, &root) != nil {
		return -1
	}
	grouped := countGroupedItems(asSlice(root["groups"]))
	if grouped < 0 {
		return -1
	}
	return len(asSlice(root["ungrouped"])) + grouped
}

func countGroupedItems(groups []any) int {
	count := 0
	for _, raw := range groups {
		group, ok := raw.(map[string]any)
		if !ok {
			return -1
		}
		nested := countGroupedItems(asSlice(group["groups"]))
		if nested < 0 {
			return -1
		}
		count += len(asSlice(group["items"])) + nested
	}
	return count
}

func asSlice(value any) []any {
	if value == nil {
		return nil
	}
	items, _ := value.([]any)
	return items
}
