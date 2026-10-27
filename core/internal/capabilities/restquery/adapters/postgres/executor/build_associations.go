package executor

import (
	"bytes"
	"sort"

	"github.com/google/uuid"
)

func buildAssociations(
	sourceIDs []uuid.UUID,
	edges []relationEdge,
) (map[uuid.UUID]map[uuid.UUID]struct{}, []uuid.UUID, error) {
	if sourceIDs == nil || len(sourceIDs) < 1 || len(sourceIDs) > 1000 || edges == nil {
		return nil, nil, ErrContract
	}
	sources := make(map[uuid.UUID]struct{}, len(sourceIDs))
	for _, source := range sourceIDs {
		if source == uuid.Nil || source.Variant() != uuid.RFC4122 {
			return nil, nil, ErrContract
		}
		if _, duplicate := sources[source]; duplicate {
			return nil, nil, ErrContract
		}
		sources[source] = struct{}{}
	}
	result := make(map[uuid.UUID]map[uuid.UUID]struct{})
	slots := make(map[uuid.UUID]map[int16]struct{})
	targets := make(map[uuid.UUID]struct{})
	for index, edge := range edges {
		if edge.validate() != nil || !edgeSourceAllowed(edge.source, sources) ||
			index > 0 && compareRelationEdges(edges[index-1], edge) >= 0 {
			return nil, nil, ErrContract
		}
		if result[edge.source] == nil {
			result[edge.source] = make(map[uuid.UUID]struct{})
			slots[edge.source] = make(map[int16]struct{})
		}
		if _, duplicate := result[edge.source][edge.target]; duplicate {
			return nil, nil, ErrContract
		}
		if _, duplicate := slots[edge.source][edge.slot]; duplicate {
			return nil, nil, ErrContract
		}
		result[edge.source][edge.target] = struct{}{}
		slots[edge.source][edge.slot] = struct{}{}
		targets[edge.target] = struct{}{}
	}
	candidates := make([]uuid.UUID, 0, len(targets))
	for target := range targets {
		candidates = append(candidates, target)
	}
	sort.Slice(candidates, func(left, right int) bool {
		return bytes.Compare(candidates[left][:], candidates[right][:]) < 0
	})
	return result, candidates, nil
}

func edgeSourceAllowed(
	source uuid.UUID,
	allowed map[uuid.UUID]struct{},
) bool {
	_, found := allowed[source]
	return found
}

func compareRelationEdges(left, right relationEdge) int {
	if order := bytes.Compare(left.source[:], right.source[:]); order != 0 {
		return order
	}
	if left.slot < right.slot {
		return -1
	}
	if left.slot > right.slot {
		return 1
	}
	return bytes.Compare(left.target[:], right.target[:])
}
