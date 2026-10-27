package documents

import (
	"context"
	"net/http"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	downloadapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/download"
)

func (runtime *HTTPRuntime) dispatchDownload(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	if !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	itemID, itemErr := parseHTTPID(request.ResourceID, false)
	versionID, versionErr := parseHTTPID(request.SecondaryID, true)
	if itemErr != nil || versionErr != nil ||
		(request.Operation == downloadapp.OperationVersion) != versionID.Valid() {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	runner := httpScopeRunner{runtime: runtime, access: request.Access}
	result, err := runtime.dependencies.Downloads.Open(ctx, runner,
		downloadapp.Request{ItemID: itemID, VersionID: versionID})
	if err != nil {
		return documentshttp.Response{}, mapHTTPApplicationError(err)
	}
	return documentshttp.Response{Status: http.StatusOK,
		Binary: &documentshttp.Binary{Filename: result.Filename, MIME: result.MIME,
			Bytes: result.Bytes, SHA256: result.SHA256, Content: result.Content}}, nil
}
