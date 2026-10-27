package dashboard

import (
	"encoding/hex"
	"encoding/json"

	dashboardapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
)

type widgetJSON struct {
	WidgetID          string `json:"widget_id"`
	Position          uint16 `json:"position"`
	QueryVersionID    string `json:"query_version_id"`
	ParameterDocument string `json:"parameter_document"`
	OutputDocument    string `json:"output_document"`
}

func encodeWidgets(values []dashboardapp.Widget) ([]byte, error) {
	if len(values) > 30 {
		return nil, dashboardapp.ErrInvalid
	}
	result := make([]widgetJSON, 0, len(values))
	for _, value := range values {
		parameters, parametersErr := dashboardapp.NewDocument(value.ParameterDocument)
		output, outputErr := dashboardapp.NewDocument(value.OutputDocument)
		if !validUUIDv7(value.WidgetID) || !validUUIDv7(value.QueryVersionID) ||
			parametersErr != nil || outputErr != nil || value.Position > 29 {
			return nil, dashboardapp.ErrInvalid
		}
		result = append(result, widgetJSON{WidgetID: value.WidgetID.String(),
			Position: value.Position, QueryVersionID: value.QueryVersionID.String(),
			ParameterDocument: hex.EncodeToString(parameters.CanonicalJSON()),
			OutputDocument:    hex.EncodeToString(output.CanonicalJSON())})
	}
	encoded, err := json.Marshal(result)
	if err != nil {
		return nil, dashboardapp.ErrUnavailable
	}
	return encoded, nil
}

func decodeWidgets(encoded []byte) ([]dashboardapp.Widget, error) {
	var values []widgetJSON
	if len(encoded) == 0 || json.Unmarshal(encoded, &values) != nil || len(values) > 30 {
		return nil, dashboardapp.ErrUnavailable
	}
	result := make([]dashboardapp.Widget, 0, len(values))
	for _, value := range values {
		widgetID, widgetErr := parseUUIDv7(value.WidgetID)
		queryID, queryErr := parseUUIDv7(value.QueryVersionID)
		parameters, parametersErr := hex.DecodeString(value.ParameterDocument)
		output, outputErr := hex.DecodeString(value.OutputDocument)
		parameterDocument, parameterDocumentErr := dashboardapp.NewDocument(parameters)
		outputDocument, outputDocumentErr := dashboardapp.NewDocument(output)
		if widgetErr != nil || queryErr != nil || parametersErr != nil || outputErr != nil ||
			parameterDocumentErr != nil || outputDocumentErr != nil || value.Position > 29 {
			return nil, dashboardapp.ErrUnavailable
		}
		result = append(result, dashboardapp.Widget{WidgetID: widgetID,
			QueryVersionID: queryID, Position: value.Position,
			ParameterDocument: parameterDocument.CanonicalJSON(),
			OutputDocument:    outputDocument.CanonicalJSON()})
	}
	return result, nil
}
