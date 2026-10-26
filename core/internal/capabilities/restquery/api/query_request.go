package api

import "fmt"

const maximumProjectedFields = 50

type QueryMode uint8

const (
	QueryModeRecord QueryMode = iota + 1
	QueryModeAggregate
)

type RecordQueryRequestInput struct {
	Select  []string
	Where   *RequestFilter
	Sort    []RequestSort
	Page    RequestPage
	Search  *RequestSearch
	Count   CountMode
	Include []RequestInclude
}

type QueryRequest struct {
	mode         QueryMode
	selectFields []string
	where        *RequestFilter
	sorts        []RequestSort
	page         RequestPage
	search       *RequestSearch
	count        CountMode
	includes     []RequestInclude
	aggregate    *RequestAggregate
}

func NewRecordQueryRequest(input RecordQueryRequestInput) (QueryRequest, error) {
	selectFields, err := cloneUniqueRequestNames(input.Select, maximumProjectedFields)
	if err != nil {
		return QueryRequest{}, fmt.Errorf("REST Query root select: %w", err)
	}
	sorts, err := cloneRequestSorts(input.Sort)
	if err != nil {
		return QueryRequest{}, err
	}
	where, err := cloneOptionalRequestFilter(input.Where)
	if err != nil {
		return QueryRequest{}, err
	}
	search, err := cloneOptionalRequestSearch(input.Search)
	if err != nil {
		return QueryRequest{}, err
	}
	includes, err := cloneRequestIncludes(input.Include)
	if err != nil {
		return QueryRequest{}, err
	}
	projected := len(selectFields)
	for _, include := range includes {
		projected += int(include.projectedFields)
	}
	if projected > maximumProjectedFields {
		return QueryRequest{}, fmt.Errorf("REST Query projected field count exceeds its limit")
	}
	if !input.Page.valid() || !input.Count.valid() {
		return QueryRequest{}, fmt.Errorf("REST Query record mode is invalid")
	}
	request := QueryRequest{
		mode: QueryModeRecord, selectFields: selectFields, where: where,
		sorts: sorts, page: input.Page, search: search,
		count: input.Count, includes: includes,
	}
	if err := request.Validate(); err != nil {
		return QueryRequest{}, err
	}
	return request, nil
}

func (request QueryRequest) Mode() QueryMode     { return request.mode }
func (request QueryRequest) Select() []string    { return append([]string{}, request.selectFields...) }
func (request QueryRequest) Sort() []RequestSort { return append([]RequestSort{}, request.sorts...) }
func (request QueryRequest) Page() RequestPage   { return request.page }
func (request QueryRequest) Count() CountMode    { return request.count }
func (request QueryRequest) Includes() []RequestInclude {
	return cloneRequestIncludeSlice(request.includes)
}
func (request QueryRequest) Aggregate() (RequestAggregate, bool) {
	if request.aggregate == nil {
		return RequestAggregate{}, false
	}
	return request.aggregate.clone(), true
}
func (request QueryRequest) Where() (RequestFilter, bool) {
	if request.where == nil {
		return RequestFilter{}, false
	}
	return request.where.clone(), true
}
func (request QueryRequest) Search() (RequestSearch, bool) {
	if request.search == nil {
		return RequestSearch{}, false
	}
	return *request.search, true
}

func (request QueryRequest) String() string                 { return queryValueRedaction }
func (request QueryRequest) GoString() string               { return queryValueRedaction }
func (request QueryRequest) Format(state fmt.State, _ rune) { writeQueryRedaction(state) }
func (QueryRequest) MarshalJSON() ([]byte, error)           { return rejectDirectJSONMarshal() }
func (input RecordQueryRequestInput) String() string        { return queryValueRedaction }
func (input RecordQueryRequestInput) GoString() string      { return queryValueRedaction }
func (input RecordQueryRequestInput) Format(state fmt.State, _ rune) {
	writeQueryRedaction(state)
}
func (RecordQueryRequestInput) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}
func (mode QueryMode) String() string {
	if mode == QueryModeRecord {
		return "record"
	}
	if mode == QueryModeAggregate {
		return "aggregate"
	}
	return ""
}
func (QueryMode) MarshalJSON() ([]byte, error) { return rejectDirectJSONMarshal() }
