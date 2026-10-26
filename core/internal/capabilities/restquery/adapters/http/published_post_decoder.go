package httpquery

import (
	"unicode/utf8"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/published"
)

const maximumPublishedPOSTBytes = 64 * 1024

type publishedPOSTDecoder struct{}

func NewPublishedPOSTDecoder() published.POSTDecoder { return publishedPOSTDecoder{} }

func (publishedPOSTDecoder) Decode(
	document []byte, manifest published.ParameterManifest,
) (published.BoundParameters, error) {
	if err := manifest.Validate(); err != nil {
		return published.BoundParameters{}, err
	}
	if len(document) > maximumPublishedPOSTBytes {
		return published.BoundParameters{},
			publishedTransportError(restqueryv1.IssueBodySizeExceeded)
	}
	if !utf8.Valid(document) {
		return published.BoundParameters{},
			publishedTransportError(restqueryv1.IssueUTF8Invalid)
	}
	scan, err := scanJSONMembers(document)
	if err != nil {
		return published.BoundParameters{},
			publishedTransportError(restqueryv1.IssueValueNotCanonical)
	}
	if paths := scan.duplicatePaths(); len(paths) != 0 {
		collector := issueCollector{}
		for _, path := range paths {
			collector.append(path, restqueryv1.IssueMemberDuplicate)
		}
		return published.BoundParameters{}, queryError(
			restqueryv1.ErrorQueryDuplicateMember, collector.issues,
		)
	}
	collector := publishedIssueCollector{}
	root, valid := decodePublishedPOSTRoot(document, &collector)
	if !valid {
		return published.BoundParameters{}, collector.queryError()
	}
	parameters := decodePublishedPOSTParameters(root, manifest, &collector)
	page := decodePublishedPOSTPage(root, &collector)
	if collector.hasIssues() {
		return published.BoundParameters{}, collector.queryError()
	}
	return published.BindPOST(manifest, parameters, page)
}
