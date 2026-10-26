package resolved

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type RecordQuery struct {
	projections       []Projection
	where             *Filter
	sorts             []Sort
	search            []SearchTerm
	page              Page
	count             api.CountMode
	countPath         RequestPath
	exactCountEnabled bool
	includes          []Include
	validated         bool
}

func NewRecordQuery(
	projections []Projection,
	where *Filter,
	sorts []Sort,
	search []SearchTerm,
	page Page,
	count api.CountMode,
	countPath RequestPath,
	exactCountEnabled bool,
	includes []Include,
) (RecordQuery, error) {
	query := RecordQuery{
		projections: append([]Projection{}, projections...), sorts: append([]Sort{}, sorts...),
		search: cloneSearchTerms(search), page: page, count: count, countPath: countPath,
		exactCountEnabled: exactCountEnabled,
		includes:          cloneIncludes(includes), validated: true,
	}
	if where != nil {
		owned := where.clone()
		query.where = &owned
	}
	if err := query.Validate(); err != nil {
		return RecordQuery{}, err
	}
	return query, nil
}

func (query RecordQuery) Validate() error {
	if !query.validated || query.projections == nil || query.sorts == nil ||
		query.search == nil || query.includes == nil || query.page.Validate() != nil ||
		query.count.String() == "" || query.countPath.Validate() != nil ||
		!validProjectionSet(query.projections, true) {
		return fmt.Errorf("resolved REST Query record query is invalid")
	}
	visible, hidden := 0, 0
	for _, projection := range query.projections {
		if projection.hidden {
			hidden++
		} else {
			visible++
		}
	}
	if visible < 1 || visible > 50 || hidden > 1 || !validRecordSorts(query.sorts) ||
		!validRecordSearch(query.search) || !validExactCount(query) {
		return fmt.Errorf("resolved REST Query record query is invalid")
	}
	if query.where != nil && query.where.Validate() != nil {
		return fmt.Errorf("resolved REST Query record query is invalid")
	}
	for _, include := range query.includes {
		if include.Validate() != nil {
			return fmt.Errorf("resolved REST Query record query is invalid")
		}
	}
	return nil
}

func (query RecordQuery) CountPath() (RequestPath, error) {
	if err := query.Validate(); err != nil {
		return RequestPath{}, err
	}
	return query.countPath, nil
}

func (query RecordQuery) Projections() ([]Projection, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return append([]Projection{}, query.projections...), nil
}

func (query RecordQuery) Where() (Filter, bool, error) {
	if err := query.Validate(); err != nil {
		return Filter{}, false, err
	}
	if query.where == nil {
		return Filter{}, false, nil
	}
	return query.where.clone(), true, nil
}

func (query RecordQuery) Sorts() ([]Sort, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return append([]Sort{}, query.sorts...), nil
}

func (query RecordQuery) Search() ([]SearchTerm, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return cloneSearchTerms(query.search), nil
}

func (query RecordQuery) Page() (Page, error) {
	if err := query.Validate(); err != nil {
		return Page{}, err
	}
	return query.page, nil
}

func (query RecordQuery) Count() (api.CountMode, error) {
	if err := query.Validate(); err != nil {
		return 0, err
	}
	return query.count, nil
}

func (query RecordQuery) Includes() ([]Include, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return cloneIncludes(query.includes), nil
}

func (RecordQuery) String() string                       { return valueRedaction }
func (RecordQuery) GoString() string                     { return valueRedaction }
func (query RecordQuery) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (RecordQuery) MarshalJSON() ([]byte, error)         { return rejectDirectJSONMarshal() }
