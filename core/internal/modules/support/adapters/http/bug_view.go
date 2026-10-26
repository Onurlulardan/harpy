package supporthttp

import "github.com/Onurlulardan/harpy/core/internal/modules/support/domain"

func bugView(report domain.BugReport) map[string]any {
	view := map[string]any{
		"id": report.ID.String(), "tenant_id": report.TenantID.String(),
		"tenant_slug": report.TenantSlug, "tenant_name": report.TenantName,
		"reporter_id": report.ReporterID.String(), "reporter_name": report.ReporterName,
		"reporter_email": report.ReporterEmail, "title": report.Title,
		"description": report.Description, "category": report.Category,
		"severity": report.Severity, "priority": report.Priority,
		"page_url": report.PageURL, "steps_to_reproduce": report.StepsToReproduce,
		"status": report.Status, "resolution": report.Resolution,
		"admin_notes": report.AdminNotes, "revision": report.Revision,
		"created_at": wireTime(report.CreatedAt), "updated_at": wireTime(report.UpdatedAt),
	}
	if report.ResolvedAt != nil {
		view["resolved_at"] = wireTime(*report.ResolvedAt)
	} else {
		view["resolved_at"] = nil
	}
	if report.ClosedAt != nil {
		view["closed_at"] = wireTime(*report.ClosedAt)
	} else {
		view["closed_at"] = nil
	}
	return view
}

func bugPageView(page domain.BugPage) map[string]any {
	items := make([]any, len(page.Items))
	for index := range page.Items {
		items[index] = bugView(page.Items[index])
	}
	return pagedView(items, page.Page, page.Total)
}

func bugStatsView(stats domain.BugStats) map[string]any {
	return map[string]any{"total": stats.Total, "open": stats.Open,
		"in_progress": stats.InProgress, "resolved": stats.Resolved,
		"closed": stats.Closed, "critical_open": stats.CriticalOpen,
		"by_category": stats.ByCategory, "by_severity": stats.BySeverity}
}
