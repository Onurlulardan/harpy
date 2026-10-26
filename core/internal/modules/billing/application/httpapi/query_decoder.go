package httpapi

import (
	"errors"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	restqueryapi "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/httpdecode"
)

type QueryDecoder struct {
	get  restqueryapi.GETDecoder
	post restqueryapi.POSTDecoder
}

func NewQueryDecoder() QueryDecoder {
	return QueryDecoder{get: httpdecode.GET(), post: httpdecode.POST()}
}

func (decoder QueryDecoder) DecodeGET(raw string) (restqueryapi.QueryRequest, error) {
	return decoder.get.Decode(raw)
}

func (decoder QueryDecoder) DecodePOST(document []byte) (restqueryapi.QueryRequest, error) {
	return decoder.post.Decode(document)
}

func QueryErrorDetails(
	err error,
) (int, string, []runtimev1.Issue, bool) {
	var query restqueryapi.QueryError
	if !errors.As(err, &query) {
		return 0, "", nil, false
	}
	status, valid := query.HTTPStatus()
	return status, string(query.Code()), query.Issues(), valid
}
