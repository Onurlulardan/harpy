package application

import "github.com/Onurlulardan/harpy/core/internal/modules/support/domain"

func reporterView(report domain.BugReport) domain.BugReport {
	report.AdminNotes = ""
	return report
}
