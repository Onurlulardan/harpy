package licenseauthority

import (
	"encoding/json"
	"sort"

	licenseapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/licenseauthority"
)

type snapshotDocuments struct {
	features, limits, projection []byte
}

type projectionWire struct {
	EntitlementKey string `json:"entitlement_key"`
	ValueType      string `json:"value_type"`
	BooleanValue   *bool  `json:"boolean_value"`
	IntegerValue   *int64 `json:"integer_value"`
	Unlimited      bool   `json:"unlimited"`
}

func encodeSnapshot(snapshot licenseapp.VerifiedSnapshot) (snapshotDocuments, error) {
	features := snapshot.Features()
	featureDocument, err := json.Marshal(map[string]bool{
		"formBuilder": features.FormBuilder, "workflowBuilder": features.WorkflowBuilder,
		"apiBuilder": features.APIBuilder, "dashboardBuilder": features.DashboardBuilder,
		"aiAssistant": features.AIAssistant,
	})
	if err != nil {
		return snapshotDocuments{}, err
	}
	limits := snapshot.Limits()
	limitDocument, err := json.Marshal(map[string]int64{
		"maxUsers": limits.MaxUsers, "maxForms": limits.MaxForms,
		"maxWorkflows": limits.MaxWorkflows, "maxApiEndpoints": limits.MaxAPIEndpoints,
		"maxDashboards": limits.MaxDashboards, "maxTenants": limits.MaxTenants,
	})
	if err != nil {
		return snapshotDocuments{}, err
	}
	projectionDocument, err := encodeProjection(licenseapp.ProjectEntitlements(snapshot))
	return snapshotDocuments{
		features: featureDocument, limits: limitDocument, projection: projectionDocument,
	}, err
}

func encodeProjection(projection licenseapp.EntitlementProjection) ([]byte, error) {
	keys := make([]string, 0, len(projection.Boolean)+len(projection.Limits))
	for key := range projection.Boolean {
		keys = append(keys, key)
	}
	for key := range projection.Limits {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	wire := make([]projectionWire, 0, len(keys))
	for _, key := range keys {
		if value, exists := projection.Boolean[key]; exists {
			copied := value
			wire = append(wire, projectionWire{
				EntitlementKey: key, ValueType: "boolean", BooleanValue: &copied,
			})
			continue
		}
		value := projection.Limits[key]
		grant := projectionWire{EntitlementKey: key, ValueType: "integer", Unlimited: value.Unlimited}
		if !value.Unlimited {
			copied := value.Value
			grant.IntegerValue = &copied
		}
		wire = append(wire, grant)
	}
	return json.Marshal(wire)
}
