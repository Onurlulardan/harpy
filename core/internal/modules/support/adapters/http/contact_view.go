package supporthttp

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func contactView(message domain.ContactMessage) map[string]any {
	view := map[string]any{
		"id": message.ID.String(), "name": message.Name, "email": message.Email,
		"company": message.Company, "phone": message.Phone,
		"employee_count": message.EmployeeCount, "message": message.Message,
		"contact_type": message.ContactType, "status": message.Status,
		"source": message.Source, "notes": message.Notes,
		"revision": message.Revision, "created_at": wireTime(message.CreatedAt),
		"updated_at": wireTime(message.UpdatedAt),
	}
	if message.RepliedAt != nil {
		view["replied_at"] = wireTime(*message.RepliedAt)
	} else {
		view["replied_at"] = nil
	}
	if message.RepliedBy != nil {
		view["replied_by"] = message.RepliedBy.String()
	} else {
		view["replied_by"] = nil
	}
	return view
}

func contactPageView(page domain.ContactPage) map[string]any {
	items := make([]any, len(page.Items))
	for index := range page.Items {
		items[index] = contactView(page.Items[index])
	}
	return pagedView(items, page.Page, page.Total)
}

func contactStatsView(stats domain.ContactStats) map[string]int64 {
	return map[string]int64{"total": stats.Total, "new": stats.New, "read": stats.Read,
		"replied": stats.Replied, "archived": stats.Archived, "this_month": stats.ThisMonth}
}
