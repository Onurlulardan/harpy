package workhttp

import restapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"

// EmptyQuery returns the published no-filter list query used by non-HTTP
// compatibility adapters without importing the REST Query transport decoder.
func EmptyQuery() (restapi.QueryRequest, error) {
	page, err := restapi.NewRequestPage(50, nil)
	if err != nil {
		return restapi.QueryRequest{}, err
	}
	return restapi.NewRecordQueryRequest(restapi.RecordQueryRequestInput{
		Select: []string{}, Sort: []restapi.RequestSort{}, Page: page,
		Count: restapi.CountModeNone, Include: []restapi.RequestInclude{},
	})
}
