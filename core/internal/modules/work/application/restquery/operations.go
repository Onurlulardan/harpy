package restquery

func operationFor(resource string) string {
	switch resource {
	case "boards":
		return "listWorkBoards"
	case "tasks":
		return "listWorkTasks"
	case "comments":
		return "listWorkComments"
	case "calendar":
		return "listWorkCalendarEvents"
	case "inbox":
		return "listWorkInbox"
	case "calendar-unified":
		return "listWorkCalendar"
	default:
		return ""
	}
}

func workQueryResources() []string {
	return []string{
		"boards", "tasks", "comments", "calendar", "inbox", "calendar-unified",
	}
}
