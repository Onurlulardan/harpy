package resolved

import "fmt"

type AggregateQuery struct {
	groups    []AggregateGroup
	measures  []AggregateMeasure
	where     *Filter
	search    []SearchTerm
	having    *Having
	sorts     []AggregateSort
	limit     int
	validated bool
}

func NewAggregateQuery(
	groups []AggregateGroup,
	measures []AggregateMeasure,
	where *Filter,
	search []SearchTerm,
	having *Having,
	sorts []AggregateSort,
	limit int,
) (AggregateQuery, error) {
	query := AggregateQuery{
		groups:   append([]AggregateGroup{}, groups...),
		measures: append([]AggregateMeasure{}, measures...),
		search:   cloneSearchTerms(search), sorts: append([]AggregateSort{}, sorts...),
		limit: limit, validated: true,
	}
	if where != nil {
		owned := where.clone()
		query.where = &owned
	}
	if having != nil {
		owned := having.clone()
		query.having = &owned
	}
	if err := query.Validate(); err != nil {
		return AggregateQuery{}, err
	}
	return query, nil
}

func (query AggregateQuery) Validate() error {
	if !query.validated || query.groups == nil || query.measures == nil ||
		query.search == nil || query.sorts == nil || len(query.groups) > 3 ||
		len(query.groups)+len(query.measures) < 1 ||
		len(query.measures) > 8 || len(query.search) > 16 || len(query.sorts) > 6 ||
		query.limit < 1 || query.limit > 500 {
		return fmt.Errorf("resolved REST Query aggregate query is invalid")
	}
	if !validAggregateDeclarations(query.groups, query.measures) ||
		!validAggregateFilters(query.where, query.search) ||
		!validAggregateHaving(query.having, query.groups, query.measures) ||
		!validAggregateSorts(query.sorts, len(query.groups), len(query.measures)) ||
		!validAggregatePredicateBudget(query.where, query.having) {
		return fmt.Errorf("resolved REST Query aggregate query is invalid")
	}
	return nil
}

func validAggregatePredicateBudget(where *Filter, having *Having) bool {
	nodes, literals := 0, 0
	if where != nil {
		metrics := measureFilter(*where, 0)
		if !metrics.valid {
			return false
		}
		nodes += metrics.nodes
		literals += metrics.literals
	}
	if having != nil {
		metrics := measureHaving(*having, 0)
		if !metrics.valid {
			return false
		}
		nodes += metrics.nodes
		literals += metrics.literals
	}
	return nodes <= 100 && literals <= 1000
}

func validAggregateDeclarations(
	groups []AggregateGroup, measures []AggregateMeasure,
) bool {
	groupKeys := make(map[string]struct{}, len(groups))
	measureKeys := make(map[string]struct{}, len(measures))
	for _, group := range groups {
		if group.Validate() != nil || !uniqueOutputKey(group.key, groupKeys) {
			return false
		}
	}
	for _, measure := range measures {
		if measure.Validate() != nil || !uniqueOutputKey(measure.key, measureKeys) {
			return false
		}
	}
	return true
}

func uniqueOutputKey(key OutputKey, seen map[string]struct{}) bool {
	value, err := key.Value()
	if err != nil {
		return false
	}
	if _, found := seen[value]; found {
		return false
	}
	seen[value] = struct{}{}
	return true
}

func validAggregateFilters(where *Filter, search []SearchTerm) bool {
	if where != nil && where.Validate() != nil {
		return false
	}
	for _, term := range search {
		if term.Validate() != nil {
			return false
		}
	}
	return true
}

func (AggregateQuery) String() string                       { return valueRedaction }
func (AggregateQuery) GoString() string                     { return valueRedaction }
func (query AggregateQuery) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (AggregateQuery) MarshalJSON() ([]byte, error)         { return rejectDirectJSONMarshal() }
