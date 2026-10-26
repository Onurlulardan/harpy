package dashboardhttp

import (
	"encoding/hex"
	"encoding/json"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

type widgetView struct {
	WidgetID       string          `json:"widget_id"`
	QueryVersionID string          `json:"query_version_id"`
	Position       uint16          `json:"position"`
	Parameters     json.RawMessage `json:"parameters"`
	Output         json.RawMessage `json:"output"`
}

type view struct {
	DashboardID     string          `json:"dashboard_id"`
	VersionID       string          `json:"version_id"`
	VersionNumber   uint64          `json:"version_number"`
	RootRevision    uint64          `json:"root_revision"`
	VersionRevision uint64          `json:"version_revision"`
	Lifecycle       string          `json:"lifecycle"`
	Visibility      string          `json:"visibility"`
	Published       bool            `json:"published"`
	Name            string          `json:"name"`
	Layout          json.RawMessage `json:"layout"`
	LayoutSHA256    string          `json:"layout_sha256"`
	ManifestSHA256  string          `json:"manifest_sha256"`
	Widgets         []widgetView    `json:"widgets"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}

func present(value dashboardapp.Snapshot) view {
	items := make([]widgetView, len(value.Widgets))
	for index, widget := range value.Widgets {
		items[index] = widgetView{WidgetID: widget.WidgetID.String(), QueryVersionID: widget.QueryVersionID.String(),
			Position: widget.Position, Parameters: widget.ParameterDocument, Output: widget.OutputDocument}
	}
	return view{DashboardID: value.DashboardID.String(), VersionID: value.VersionID.String(),
		VersionNumber: value.VersionNumber, RootRevision: value.RootRevision, VersionRevision: value.VersionRevision,
		Lifecycle: string(value.Lifecycle), Visibility: string(value.Visibility), Published: value.Published,
		Name: value.Name, Layout: value.LayoutDocument, LayoutSHA256: hex.EncodeToString(value.LayoutSHA256[:]),
		ManifestSHA256: hex.EncodeToString(value.ManifestSHA256[:]), Widgets: items,
		CreatedAt: value.CreatedAt.Format(httptransport.MicrosecondTime),
		UpdatedAt: value.UpdatedAt.Format(httptransport.MicrosecondTime)}
}

func presentPage(value dashboardapp.Page) map[string]any {
	items := make([]view, len(value.Items))
	for index := range value.Items {
		items[index] = present(value.Items[index])
	}
	var after any
	if value.After != nil {
		after = value.After.String()
	}
	return map[string]any{"items": items, "after": after}
}
