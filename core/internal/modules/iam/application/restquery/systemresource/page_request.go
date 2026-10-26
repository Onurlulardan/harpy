package systemresourcequery

import restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

func NewPageRequest(
	base Request,
	selectFields []string,
	size int,
	after *string,
) (Request, error) {
	value, err := base.Value()
	if err != nil || value.Mode() != restquery.QueryModeRecord ||
		len(value.Includes()) != 0 || value.Count() != restquery.CountModeNone {
		return Request{}, ErrContract
	}
	page, err := restquery.NewRequestPage(size, after)
	if err != nil {
		return Request{}, ErrContract
	}
	input := restquery.RecordQueryRequestInput{
		Select: append([]string(nil), selectFields...),
		Sort:   value.Sort(), Page: page, Count: restquery.CountModeNone,
		Include: []restquery.RequestInclude{},
	}
	if where, present := value.Where(); present {
		input.Where = &where
	}
	if search, present := value.Search(); present {
		input.Search = &search
	}
	paged, err := restquery.NewRecordQueryRequest(input)
	if err != nil {
		return Request{}, ErrContract
	}
	return newRequest(paged, base.Transport())
}
