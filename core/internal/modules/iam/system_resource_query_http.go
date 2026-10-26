package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
)

// DecodeSystemResourceGETQuery decodes the canonical V2 GET dialect at the IAM
// edge and returns only the opaque request accepted by the resource port.
func DecodeSystemResourceGETQuery(rawQuery string) (
	publicresource.QueryRequest,
	error,
) {
	request, err := queryapp.DecodeGET(rawQuery)
	if err != nil {
		return nil, err
	}
	request, err = queryapp.EnsureGETSelect(
		request, []string{"id", "revision"},
	)
	if err != nil {
		return nil, err
	}
	value, err := request.Value()
	if err != nil {
		return nil, err
	}
	return value, nil
}

type SystemResourceQueryFailure = queryapp.Failure
