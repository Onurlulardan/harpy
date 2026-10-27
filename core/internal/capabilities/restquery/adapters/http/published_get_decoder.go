package httpquery

import (
	"errors"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/published"
)

type publishedGETDecoder struct{}

func NewPublishedGETDecoder() published.GETDecoder { return publishedGETDecoder{} }

func (publishedGETDecoder) Decode(
	rawQuery string, manifest published.ParameterManifest,
) (published.BoundParameters, error) {
	if err := manifest.Validate(); err != nil {
		return published.BoundParameters{}, err
	}
	if len(rawQuery) > maximumGETQueryBytes {
		return published.BoundParameters{},
			publishedTransportError(restqueryv1.IssueQuerySizeExceeded)
	}
	leaves, err := decodeGETLeaves(rawQuery)
	if err != nil {
		return published.BoundParameters{}, publishedGETLeafError(err)
	}
	collector := publishedIssueCollector{}
	document := decodePublishedGETDocument(leaves, &collector)
	if collector.hasIssues() {
		return published.BoundParameters{}, collector.queryError()
	}
	parameters := normalizePublishedGETParameters(document, manifest, &collector)
	page := normalizePublishedGETPage(document, &collector)
	if collector.hasIssues() {
		return published.BoundParameters{}, collector.queryError()
	}
	return published.BindGET(manifest, parameters, page)
}

func publishedGETLeafError(err error) error {
	var queryError api.QueryError
	if !errors.As(err, &queryError) ||
		queryError.Code() != restqueryv1.ErrorQueryInvalid {
		return err
	}
	result, resultErr := api.NewQueryError(
		restqueryv1.ErrorQueryParameterInvalid, queryError.Issues(),
	)
	if resultErr == nil {
		return result
	}
	return err
}
