package recordkernel

import (
	"slices"

	"github.com/google/uuid"
)

func topologicalOrder(
	nodes []uuid.UUID,
	dependencies map[uuid.UUID]map[uuid.UUID]struct{},
) ([]uuid.UUID, error) {
	nodeSet := make(map[uuid.UUID]struct{}, len(nodes))
	for _, node := range nodes {
		nodeSet[node] = struct{}{}
	}
	indegree := make(map[uuid.UUID]int, len(nodes))
	dependents := make(map[uuid.UUID][]uuid.UUID, len(nodes))
	for _, node := range nodes {
		for dependency := range dependencies[node] {
			if _, internal := nodeSet[dependency]; !internal {
				continue
			}
			indegree[node]++
			dependents[dependency] = append(dependents[dependency], node)
		}
	}
	ready := make([]uuid.UUID, 0, len(nodes))
	for _, node := range nodes {
		if indegree[node] == 0 {
			ready = append(ready, node)
		}
	}
	slices.SortFunc(ready, compareUUID)
	result := make([]uuid.UUID, 0, len(nodes))
	for len(ready) > 0 {
		node := ready[0]
		ready = ready[1:]
		result = append(result, node)
		for _, dependent := range dependents[node] {
			indegree[dependent]--
			if indegree[dependent] == 0 {
				ready = append(ready, dependent)
				slices.SortFunc(ready, compareUUID)
			}
		}
	}
	if len(result) != len(nodes) {
		return nil, ErrCatalogInvalid
	}
	return result, nil
}
