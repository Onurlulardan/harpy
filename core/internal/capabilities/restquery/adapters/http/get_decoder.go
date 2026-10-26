package httpquery

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

const maximumGETQueryBytes = 16 * 1024

type getDecoder struct{}

func NewGETDecoder() api.GETDecoder { return getDecoder{} }

func (getDecoder) Decode(rawQuery string) (api.QueryRequest, error) {
	if len(rawQuery) > maximumGETQueryBytes {
		return api.QueryRequest{}, getTransportError(restqueryv1.IssueQuerySizeExceeded)
	}
	leaves, err := decodeGETLeaves(rawQuery)
	if err != nil {
		return api.QueryRequest{}, err
	}
	collector := issueCollector{}
	document := decodeGETDocument(leaves, &collector)
	if collector.hasIssues() {
		return api.QueryRequest{}, collector.queryError()
	}
	request, valid := normalizeGETDocument(document, &collector)
	if !valid || collector.hasIssues() {
		return api.QueryRequest{}, collector.queryError()
	}
	return request, nil
}

func getTransportError(code restqueryv1.IssueCode) error {
	return singleQueryError(restqueryv1.ErrorQueryTransportInvalid, "", code)
}
