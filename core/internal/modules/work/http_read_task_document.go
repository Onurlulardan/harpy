package work

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

const maxLegacyBoardTasks = 50_000

type legacyTaskReadDocument struct {
	root   []map[string]json.RawMessage
	tasks  []legacyTaskReadItem
	boards []legacyBoardTaskGroups
}

type legacyTaskReadItem struct {
	id     ports.ID
	kind   ports.TaskKind
	values map[string]json.RawMessage
}

type legacyBoardTaskGroups struct {
	rootIndex int
	groups    map[string][]map[string]json.RawMessage
}

func parseLegacyTaskReadDocument(
	document json.RawMessage, boardDetail bool,
) (*legacyTaskReadDocument, error) {
	trimmed := bytes.TrimSpace(document)
	if len(trimmed) == 0 || trimmed[0] != '[' || !json.Valid(trimmed) {
		return nil, ports.ErrUnavailable
	}
	result := &legacyTaskReadDocument{}
	if json.Unmarshal(trimmed, &result.root) != nil ||
		boardDetail && len(result.root) > 1 {
		return nil, ports.ErrUnavailable
	}
	seen := make(map[ports.ID]struct{})
	for index, record := range result.root {
		if !boardDetail {
			item, err := parseLegacyTaskReadItem(record, seen)
			if err != nil {
				return nil, err
			}
			result.tasks = append(result.tasks, item)
			continue
		}
		rawGroups, ok := record["tasks"]
		if !ok {
			return nil, ports.ErrUnavailable
		}
		var groups map[string][]map[string]json.RawMessage
		if json.Unmarshal(rawGroups, &groups) != nil || groups == nil {
			return nil, ports.ErrUnavailable
		}
		keys := make([]string, 0, len(groups))
		for key := range groups {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			for _, task := range groups[key] {
				item, err := parseLegacyTaskReadItem(task, seen)
				if err != nil {
					return nil, err
				}
				result.tasks = append(result.tasks, item)
			}
		}
		result.boards = append(result.boards, legacyBoardTaskGroups{
			rootIndex: index, groups: groups})
	}
	if len(result.tasks) > maxLegacyBoardTasks {
		return nil, ports.ErrUnavailable
	}
	return result, nil
}

func parseLegacyTaskReadItem(
	values map[string]json.RawMessage, seen map[ports.ID]struct{},
) (legacyTaskReadItem, error) {
	var rawID, rawKind string
	if values == nil || json.Unmarshal(values["id"], &rawID) != nil ||
		json.Unmarshal(values["task_type"], &rawKind) != nil {
		return legacyTaskReadItem{}, ports.ErrUnavailable
	}
	id, err := ports.ParseID(rawID)
	kind := ports.TaskKind(rawKind)
	if err != nil || !kind.Valid() {
		return legacyTaskReadItem{}, ports.ErrUnavailable
	}
	if _, duplicate := seen[id]; duplicate {
		return legacyTaskReadItem{}, ports.ErrUnavailable
	}
	seen[id] = struct{}{}
	return legacyTaskReadItem{id: id, kind: kind, values: values}, nil
}

func (document *legacyTaskReadDocument) taskIDs() []ports.ID {
	result := make([]ports.ID, len(document.tasks))
	for index, task := range document.tasks {
		result[index] = task.id
	}
	return result
}

func (document *legacyTaskReadDocument) marshal() (json.RawMessage, error) {
	for _, board := range document.boards {
		encoded, err := json.Marshal(board.groups)
		if err != nil {
			return nil, ports.ErrUnavailable
		}
		document.root[board.rootIndex]["tasks"] = encoded
	}
	encoded, err := json.Marshal(document.root)
	if err != nil {
		return nil, ports.ErrUnavailable
	}
	return json.RawMessage(encoded), nil
}
