package v1adoption

func mapStatus(value string) (string, bool) {
	switch value {
	case "TODO":
		return "OPEN", true
	case "IN_PROGRESS", "IN_REVIEW", "DONE", "CANCELLED":
		return value, true
	default:
		return "", false
	}
}

func mapColumnStatus(value string) (string, bool) {
	status, ok := mapStatus(value)
	return status, ok && status != "CANCELLED"
}

func mapTaskKind(value string) (string, bool) {
	switch value {
	case "CUSTOM", "FORM_FILL", "FORM_REVIEW":
		return value, true
	default:
		return "", false
	}
}

func mapRole(value string) (string, bool) {
	switch value {
	case "VIEWER", "EDITOR", "ADMIN":
		return value, true
	default:
		return "", false
	}
}
