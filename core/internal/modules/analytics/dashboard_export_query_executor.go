package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type dashboardExportQueryExecutor struct{ runtime *HTTPRuntime }

func NewDashboardExportQueryExecutor(
	dependencies PlatformQueryRuntimeDependencies,
) (DashboardExportWidgetExecutor, error) {
	values := httpPlatformValues{}
	effects, err := NewAsyncEffects(dependencies.Catalogs)
	if err != nil {
		return nil, fmt.Errorf("build Analytics export query effects: %w", err)
	}
	queries, err := buildPlatformQueryApplication(dependencies, values, effects)
	if err != nil {
		return nil, err
	}
	runtime := &HTTPRuntime{resolver: dependencies.Generations,
		coordinator: dependencies.Scopes, queries: queries}
	return &dashboardExportQueryExecutor{runtime: runtime}, nil
}

func (executor *dashboardExportQueryExecutor) Execute(
	ctx context.Context, pin exportapp.Pin, widget exportapp.PinnedWidget,
) (exportapp.WidgetResult, error) {
	if executor == nil || executor.runtime == nil || ctx == nil || ctx.Err() != nil {
		return exportapp.WidgetResult{}, exportapp.ErrUnavailable
	}
	parameters, err := mergeDashboardExportParameters(widget.Parameters, pin.Parameters)
	if err != nil {
		return exportapp.WidgetResult{}, exportapp.ErrInvalid
	}
	access, err := NewHTTPAccess(pin.ActorID, pin.TenantID, pin.TokenVersion,
		pin.AuthorizationVersion, pin.AuthorizationClock, time.Time{},
		PermissionQueryExecute, OperationQueryExecute)
	if err != nil {
		return exportapp.WidgetResult{}, exportapp.ErrUnavailable
	}
	var result queryapp.ExecutionResult
	err = executor.runtime.queryOwnerSnapshot(ctx, access, PermissionQueryExecute,
		OperationQueryExecute, func(scoped context.Context, scope platformtx.TenantScope) error {
			var executeErr error
			result, executeErr = executor.runtime.queries.ExecutePinned(scoped, scope,
				queryapp.ExecutePinnedRequest{VersionID: widget.QueryVersionID,
					Parameters: parameters})
			return executeErr
		})
	if err != nil {
		return exportapp.WidgetResult{}, err
	}
	rows, err := dashboardExportResultRows(result.Body)
	if err != nil {
		return exportapp.WidgetResult{}, exportapp.ErrUnavailable
	}
	return exportapp.WidgetResult{Body: append([]byte(nil), result.Body...), Rows: rows}, nil
}

func mergeDashboardExportParameters(base, override []byte) ([]byte, error) {
	values := map[string]json.RawMessage{}
	for _, document := range [][]byte{base, override} {
		if len(document) == 0 {
			document = []byte(`{}`)
		}
		var next map[string]json.RawMessage
		if json.Unmarshal(document, &next) != nil || next == nil {
			return nil, exportapp.ErrInvalid
		}
		for key, value := range next {
			values[key] = append(json.RawMessage(nil), value...)
		}
	}
	wire, err := json.Marshal(values)
	if err != nil {
		return nil, exportapp.ErrUnavailable
	}
	return canonicaljsonv1.CanonicalizeObject(wire, 64<<10)
}

func dashboardExportResultRows(body []byte) (int64, error) {
	var document struct {
		Items []json.RawMessage `json:"items"`
	}
	if json.Unmarshal(body, &document) != nil || document.Items == nil {
		return 0, exportapp.ErrUnavailable
	}
	for _, item := range document.Items {
		if !json.Valid(item) {
			return 0, exportapp.ErrUnavailable
		}
	}
	return int64(len(document.Items)), nil
}
