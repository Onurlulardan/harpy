package dashboard

import analyticsdashboard "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dashboard"

// ParseVisibility converts the wire value through the dashboard application
// boundary so outer adapters do not depend on the domain package directly.
func ParseVisibility(value string) (analyticsdashboard.Visibility, error) {
	visibility := analyticsdashboard.Visibility(value)
	if !visibility.Valid() {
		return "", ErrInvalid
	}
	return visibility, nil
}
