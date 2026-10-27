package presentation

import "strconv"

func decimal(value int) string {
	return strconv.Itoa(value)
}

func validContainerVariant(value string) bool {
	switch value {
	case "stack", "grid", "section", "card", "tabs", "tab":
		return true
	default:
		return false
	}
}

func validTextVariant(value string) bool {
	switch value {
	case "heading", "body", "caption":
		return true
	default:
		return false
	}
}

func validInteraction(value string) bool {
	switch value {
	case "auto", "read_only", "disabled":
		return true
	default:
		return false
	}
}
