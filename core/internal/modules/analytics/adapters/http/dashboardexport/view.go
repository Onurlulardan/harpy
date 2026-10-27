package dashboardexporthttp

import (
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

type view struct {
	ExportID              string `json:"export_id"`
	DashboardID           string `json:"dashboard_id"`
	DashboardVersionID    string `json:"dashboard_version_id"`
	State                 string `json:"state"`
	Format                string `json:"format"`
	Filename              string `json:"filename,omitempty"`
	ContentType           string `json:"content_type,omitempty"`
	FailureCode           string `json:"failure_code,omitempty"`
	RowCount              int64  `json:"row_count"`
	PlaintextBytes        int64  `json:"plaintext_bytes"`
	CancellationRequested bool   `json:"cancellation_requested"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
	ExpiresAt             string `json:"expires_at"`
}

func present(value exportapp.Snapshot) view {
	return view{ExportID: value.ExportID.String(), DashboardID: value.DashboardID.String(),
		DashboardVersionID: value.DashboardVersionID.String(), State: string(value.State),
		Format: string(value.Format), Filename: value.Filename, ContentType: value.ContentType,
		FailureCode: value.FailureCode, RowCount: value.RowCount,
		PlaintextBytes:        value.PlaintextBytes,
		CancellationRequested: value.CancellationRequested,
		CreatedAt:             value.CreatedAt.Format(httptransport.MicrosecondTime),
		UpdatedAt:             value.UpdatedAt.Format(httptransport.MicrosecondTime),
		ExpiresAt:             value.ExpiresAt.Format(httptransport.MicrosecondTime)}
}
