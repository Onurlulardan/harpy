package api

import "fmt"

func (include RequestInclude) valid() bool {
	if !validRequestName(include.relation) || include.depth < 1 ||
		include.depth > maximumIncludeDepth || include.edges < 1 ||
		include.edges > maximumIncludeEdges || include.projectedFields > maximumProjectedFields ||
		include.hasLimit && (include.limit < 1 || include.limit > maximumIncludeLimit) {
		return false
	}
	if _, err := cloneUniqueRequestNames(include.selectFields, maximumProjectedFields); err != nil {
		return false
	}
	if _, err := cloneRequestSorts(include.sorts); err != nil {
		return false
	}
	if include.where != nil && !include.where.valid() {
		return false
	}
	depth, edges, projected := uint8(1), uint8(1), uint8(len(include.selectFields))
	seen := make(map[string]struct{}, len(include.children))
	for _, child := range include.children {
		if !child.valid() {
			return false
		}
		if _, duplicate := seen[child.relation]; duplicate {
			return false
		}
		seen[child.relation] = struct{}{}
		depth = max(depth, child.depth+1)
		edges += child.edges
		projected += child.projectedFields
	}
	return depth == include.depth && edges == include.edges &&
		projected == include.projectedFields && depth <= maximumIncludeDepth &&
		edges <= maximumIncludeEdges && projected <= maximumProjectedFields
}

func cloneRequestIncludes(values []RequestInclude) ([]RequestInclude, error) {
	result := make([]RequestInclude, len(values))
	seen := make(map[string]struct{}, len(values))
	edges := 0
	for index, value := range values {
		if !value.valid() {
			return nil, fmt.Errorf("REST Query include %d is invalid", index)
		}
		if _, duplicate := seen[value.relation]; duplicate {
			return nil, fmt.Errorf("REST Query include relation is duplicated")
		}
		seen[value.relation] = struct{}{}
		edges += int(value.edges)
		if edges > maximumIncludeEdges {
			return nil, fmt.Errorf("REST Query include edge count exceeds its limit")
		}
		result[index] = value.clone()
	}
	return result, nil
}

func cloneRequestIncludeSlice(values []RequestInclude) []RequestInclude {
	result := make([]RequestInclude, len(values))
	for index := range values {
		result[index] = values[index].clone()
	}
	return result
}
