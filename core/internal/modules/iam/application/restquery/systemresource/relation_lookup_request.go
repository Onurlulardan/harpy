package systemresourcequery

import (
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

const RelationLookupBatchSize = 100

func NewRelationLookupRequest(
	ids []uuid.UUID,
	fields []string,
) (publicresource.QueryRequest, error) {
	if len(ids) == 0 || len(ids) > RelationLookupBatchSize ||
		len(fields) == 0 {
		return nil, ErrContract
	}
	literals := make([]api.Literal, len(ids))
	for index, id := range ids {
		encoded, err := json.Marshal(id.String())
		if err != nil {
			return nil, ErrContract
		}
		literals[index], err = api.NewJSONLiteral(encoded)
		if err != nil {
			return nil, ErrContract
		}
	}
	where, err := api.NewRequestPredicate("id", api.OperatorIn, literals)
	if err != nil {
		return nil, ErrContract
	}
	page, err := api.NewRequestPage(len(ids), nil)
	if err != nil {
		return nil, ErrContract
	}
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: append([]string(nil), fields...), Where: &where,
		Sort: []api.RequestSort{}, Page: page,
		Count: api.CountModeNone, Include: []api.RequestInclude{},
	})
	if err != nil {
		return nil, ErrContract
	}
	return request, nil
}
