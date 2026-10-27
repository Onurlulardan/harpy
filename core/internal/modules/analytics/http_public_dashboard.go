package analytics

import (
	"bytes"
	"context"
	"encoding/json"

	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (executor *publicExecutor) executeDashboard(
	ctx context.Context, scope platformtx.TenantScope,
	route publicationapp.PublicRoute, parameters []byte,
) (publicationapp.PublicResult, error) {
	if !bytes.Equal(bytes.TrimSpace(parameters), []byte(`{}`)) {
		return publicationapp.PublicResult{}, publicationapp.ErrInvalid
	}
	dashboardID, err := executor.roots.DashboardRoot(
		ctx, scope, route.ReleaseID, route.ResourceVersionID)
	if err != nil {
		return publicationapp.PublicResult{}, publicationapp.ErrUnavailable
	}
	dashboard, err := executor.dashboards.LoadPublishedVersion(
		ctx, scope, dashboardID, route.ResourceVersionID)
	if err != nil || !dashboard.Published || len(dashboard.Widgets) == 0 ||
		len(dashboard.Widgets) > 30 {
		return publicationapp.PublicResult{}, publicationapp.ErrUnavailable
	}
	result := publicDashboardResult{Widgets: make([]publicWidgetResult, 0, len(dashboard.Widgets))}
	for _, widget := range dashboard.Widgets {
		if ctx.Err() != nil || !identityOutput(widget.OutputDocument) {
			return publicationapp.PublicResult{}, publicationapp.ErrUnavailable
		}
		queryResult, executeErr := executor.queries.Run(ctx, scope,
			route.ReleaseID, widget.QueryVersionID, widget.ParameterDocument)
		if executeErr != nil || !json.Valid(queryResult.Body) {
			return publicationapp.PublicResult{}, publicationapp.ErrUnavailable
		}
		result.Widgets = append(result.Widgets, publicWidgetResult{
			WidgetID: widget.WidgetID.String(), Body: json.RawMessage(queryResult.Body)})
	}
	body, err := json.Marshal(result)
	if err != nil || len(body) == 0 || len(body) > 1<<20 {
		return publicationapp.PublicResult{}, publicationapp.ErrUnavailable
	}
	return publicationapp.PublicResult{Body: body}, nil
}

func identityOutput(document []byte) bool {
	trimmed := bytes.TrimSpace(document)
	return bytes.Equal(trimmed, []byte(`{}`)) || bytes.Equal(trimmed,
		[]byte(`{"contract":"restquery-output.v1","mode":"identity"}`))
}
