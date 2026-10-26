package executor

import (
	"bytes"
	"sort"

	"github.com/google/uuid"
)

func (runtime *graphRuntime) selectedDepthOneSources(
	nodeIndex int,
) ([]uuid.UUID, error) {
	if runtime == nil || nodeIndex < 0 || nodeIndex >= len(runtime.nodes) ||
		runtime.nodes[nodeIndex].material.depth != 1 {
		return nil, ErrContract
	}
	selectedIDs := make(map[uuid.UUID]struct{})
	occurrences := 0
	for _, root := range runtime.root {
		for index := range runtime.nodes {
			node := runtime.nodes[index].material
			if node.depth != 1 {
				continue
			}
			selected, err := node.selectRecords(root.id)
			if err != nil {
				return nil, err
			}
			occurrences += len(selected)
			if occurrences > maximumIncludedOccurrences {
				return nil, includeResultBudgetError()
			}
			if index == nodeIndex {
				for _, record := range selected {
					selectedIDs[record.id] = struct{}{}
				}
			}
		}
	}
	return sortedUUIDSet(selectedIDs), nil
}

func sortedUUIDSet(values map[uuid.UUID]struct{}) []uuid.UUID {
	result := make([]uuid.UUID, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Slice(result, func(left, right int) bool {
		return bytes.Compare(result[left][:], result[right][:]) < 0
	})
	return result
}

func materialRecordIDs(records []materialRecord) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, len(records))
	for index, record := range records {
		if record.validate() != nil {
			return nil, ErrContract
		}
		result[index] = record.id
	}
	return result, nil
}

func (runtime *graphRuntime) clearOutgoing(parent int) {
	for _, child := range runtime.children(parent) {
		runtime.nodes[child].material.associations =
			map[uuid.UUID]map[uuid.UUID]struct{}{}
		runtime.nodes[child].candidates = []uuid.UUID{}
	}
}
