package domain

import "strings"

func FeatureForRoute(route string) string {
	value := strings.ToLower(route)
	switch {
	case strings.HasPrefix(value, "/api/auth/"):
		return "Authentication"
	case strings.Contains(value, "/form-engine/dynamic/"):
		return dynamicFormFeature(value)
	case strings.HasPrefix(value, "/api/workflow"):
		return "Workflows"
	case strings.HasPrefix(value, "/api/notifications"):
		return "Notifications"
	case strings.HasPrefix(value, "/api/documents"), strings.HasPrefix(value, "/api/files"):
		return "Documents"
	case strings.HasPrefix(value, "/api/traces"):
		return "Traces"
	case strings.HasPrefix(value, "/api/system/health"), strings.HasPrefix(value, "/api/alerts"):
		return "System Health"
	case strings.HasPrefix(value, "/api/platform/operations"):
		return "Platform Operations"
	case strings.HasPrefix(value, "/api/changelog"):
		return "Changelog"
	case strings.HasPrefix(value, "/api/analytics"), strings.HasPrefix(value, "/api/usage-analytics"):
		return "Analytics"
	case strings.HasPrefix(value, "/api/permissions"):
		return "Permissions"
	case strings.HasPrefix(value, "/api/admin"):
		return "Administration"
	case strings.HasPrefix(value, "/api/menu"), strings.HasPrefix(value, "/api/navigation"):
		return "Navigation"
	case strings.HasPrefix(value, "/api/dashboard"):
		return "Dashboards"
	case strings.HasPrefix(value, "/api/ai"), strings.HasPrefix(value, "/api/chat"):
		return "AI Chat"
	case strings.HasPrefix(value, "/api/openapi"):
		return "OpenAPI"
	case strings.HasPrefix(value, "/api/invitations"):
		return "Invitations"
	default:
		return "Other"
	}
}

func dynamicFormFeature(route string) string {
	const marker = "/form-engine/dynamic/"
	remainder := route[strings.Index(route, marker)+len(marker):]
	if slug := strings.SplitN(remainder, "/", 2)[0]; slug != "" {
		return "Form: " + slug
	}
	return "Forms"
}
