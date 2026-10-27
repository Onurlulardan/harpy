package httpquery

import (
	"unicode/utf8"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

const maximumPOSTDocumentBytes = 64 * 1024

type postDecoder struct{}

func NewPOSTDecoder() api.POSTDecoder { return postDecoder{} }

func (postDecoder) Decode(document []byte) (api.QueryRequest, error) {
	if len(document) > maximumPOSTDocumentBytes {
		return api.QueryRequest{}, singleQueryError(
			restqueryv1.ErrorQueryTransportInvalid, "", restqueryv1.IssueBodySizeExceeded,
		)
	}
	if !utf8.Valid(document) {
		return api.QueryRequest{}, singleQueryError(
			restqueryv1.ErrorQueryTransportInvalid, "", restqueryv1.IssueUTF8Invalid,
		)
	}
	scan, err := scanJSONMembers(document)
	if err != nil {
		return api.QueryRequest{}, malformedPOSTError(err)
	}
	if paths := scan.duplicatePaths(); len(paths) != 0 {
		collector := issueCollector{}
		for _, path := range paths {
			collector.append(path, restqueryv1.IssueMemberDuplicate)
		}
		return api.QueryRequest{}, queryError(
			restqueryv1.ErrorQueryDuplicateMember, collector.issues,
		)
	}
	collector := issueCollector{}
	queryDocument, valid := decodeQueryDocument(document, &collector)
	if !valid {
		return api.QueryRequest{}, collector.queryError()
	}
	request, valid := normalizeQueryDocument(queryDocument, &collector)
	if !valid || collector.hasIssues() {
		return api.QueryRequest{}, collector.queryError()
	}
	return request, nil
}

func malformedPOSTError(_ error) error {
	return singleQueryError(
		restqueryv1.ErrorQueryTransportInvalid, "", restqueryv1.IssueValueNotCanonical,
	)
}
