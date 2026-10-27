package documents

import (
	"context"
	"net/http"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	itemapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/item"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (runtime *HTTPRuntime) listHTTPVersions(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	itemID, err := parseHTTPID(request.ResourceID, false)
	if err != nil || !onlyHTTPQuery(request.Query, "after_number", "limit") {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	after, afterErr := httpPositiveUint(request.Query, "after_number")
	limit, limitErr := httpLimit(request.Query, 100)
	if afterErr != nil || limitErr != nil {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	var result itemapp.VersionPage
	err = runtime.snapshot(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var listErr error
			result, listErr = runtime.dependencies.Items.ReadVersions(scoped, scope,
				itemapp.ListVersionsRequest{ItemID: itemID,
					AfterVersionNumber: after, Limit: limit})
			return listErr
		})
	views := make([]httpVersionView, len(result.Items))
	for index, version := range result.Items {
		views[index] = httpVersion(version)
	}
	data := map[string]any{"items": views, "next_after_number": result.NextAfterNumber}
	return documentshttp.Response{Status: http.StatusOK, Data: data},
		mapHTTPApplicationError(err)
}

func (runtime *HTTPRuntime) latestHTTPVersion(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	itemID, err := parseHTTPID(request.ResourceID, false)
	if err != nil || !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	var result itemapp.Version
	err = runtime.snapshot(ctx, request.Access, request.Operation,
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var latestErr error
			result, latestErr = runtime.dependencies.Items.LatestVersion(scoped, scope, itemID)
			return latestErr
		})
	return documentshttp.Response{Status: http.StatusOK, Data: httpVersion(result)},
		mapHTTPApplicationError(err)
}
