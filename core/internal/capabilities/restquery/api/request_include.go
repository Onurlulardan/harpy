package api

import "fmt"

const (
	maximumIncludeDepth = 2
	maximumIncludeEdges = 3
	maximumIncludeLimit = 100
)

type RequestIncludeInput struct {
	Relation string
	Select   []string
	Where    *RequestFilter
	Sort     []RequestSort
	Limit    *int
	Include  []RequestInclude
}

type RequestInclude struct {
	relation        string
	selectFields    []string
	where           *RequestFilter
	sorts           []RequestSort
	limit           int
	hasLimit        bool
	children        []RequestInclude
	depth           uint8
	edges           uint8
	projectedFields uint8
}

func NewRequestInclude(input RequestIncludeInput) (RequestInclude, error) {
	if !validRequestName(input.Relation) {
		return RequestInclude{}, fmt.Errorf("REST Query include relation is invalid")
	}
	selectFields, err := cloneUniqueRequestNames(input.Select, maximumProjectedFields)
	if err != nil {
		return RequestInclude{}, fmt.Errorf("REST Query include select: %w", err)
	}
	sorts, err := cloneRequestSorts(input.Sort)
	if err != nil {
		return RequestInclude{}, err
	}
	where, err := cloneOptionalRequestFilter(input.Where)
	if err != nil {
		return RequestInclude{}, err
	}
	children, err := cloneRequestIncludes(input.Include)
	if err != nil {
		return RequestInclude{}, err
	}
	result := RequestInclude{
		relation: input.Relation, selectFields: selectFields, where: where,
		sorts: sorts, children: children, depth: 1, edges: 1,
		projectedFields: uint8(len(selectFields)),
	}
	if input.Limit != nil {
		result.limit, result.hasLimit = *input.Limit, true
	}
	for _, child := range children {
		result.depth = max(result.depth, child.depth+1)
		result.edges += child.edges
		result.projectedFields += child.projectedFields
	}
	if !result.valid() {
		return RequestInclude{}, fmt.Errorf("REST Query include structure is invalid")
	}
	return result, nil
}

func (include RequestInclude) Relation() string { return include.relation }
func (include RequestInclude) Select() []string { return append([]string{}, include.selectFields...) }
func (include RequestInclude) Sort() []RequestSort {
	return append([]RequestSort{}, include.sorts...)
}
func (include RequestInclude) Limit() (int, bool) { return include.limit, include.hasLimit }
func (include RequestInclude) Where() (RequestFilter, bool) {
	if include.where == nil {
		return RequestFilter{}, false
	}
	return include.where.clone(), true
}
func (include RequestInclude) Includes() []RequestInclude {
	return cloneRequestIncludeSlice(include.children)
}

func (include RequestInclude) String() string   { return queryValueRedaction }
func (include RequestInclude) GoString() string { return queryValueRedaction }
func (include RequestInclude) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (RequestInclude) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
func (RequestIncludeInput) String() string          { return queryValueRedaction }
func (RequestIncludeInput) GoString() string        { return queryValueRedaction }
func (RequestIncludeInput) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (RequestIncludeInput) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}

func (include RequestInclude) clone() RequestInclude {
	result := include
	result.selectFields = append([]string{}, include.selectFields...)
	result.sorts = append([]RequestSort{}, include.sorts...)
	result.children = cloneRequestIncludeSlice(include.children)
	if include.where != nil {
		where := include.where.clone()
		result.where = &where
	}
	return result
}
