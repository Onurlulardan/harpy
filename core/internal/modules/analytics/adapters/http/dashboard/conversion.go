package dashboardhttp

import (
	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
)

func widgets(values []widgetDocument) ([]dashboardapp.Widget, bool) {
	if len(values) > 30 {
		return nil, false
	}
	result := make([]dashboardapp.Widget, len(values))
	seenIDs := make(map[string]struct{}, len(values))
	seenPositions := make(map[uint16]struct{}, len(values))
	for index, value := range values {
		widgetID, widgetOK := httptransport.WireV7(value.WidgetID)
		queryVersionID, queryOK := httptransport.WireV7(value.QueryVersionID)
		if !widgetOK || !queryOK || value.Position >= 30 ||
			len(value.Parameters) == 0 || len(value.Output) == 0 {
			return nil, false
		}
		if _, exists := seenIDs[value.WidgetID]; exists {
			return nil, false
		}
		if _, exists := seenPositions[value.Position]; exists {
			return nil, false
		}
		seenIDs[value.WidgetID] = struct{}{}
		seenPositions[value.Position] = struct{}{}
		result[index] = dashboardapp.Widget{WidgetID: widgetID, QueryVersionID: queryVersionID,
			Position: value.Position, ParameterDocument: value.Parameters, OutputDocument: value.Output}
	}
	return result, true
}
