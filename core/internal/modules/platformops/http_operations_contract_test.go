package platformops

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type operationsPathContract struct {
	PathItems map[string]operationsContractPath `json:"pathItems"`
}

type operationsContractOperation struct {
	OperationID string `json:"operationId"`
	Capability  string `json:"x-harpy-capability"`
}

type operationsContractPath struct {
	Get, Post, Put, Delete *operationsContractOperation
}

func (path operationsContractPath) operations() map[string]*operationsContractOperation {
	return map[string]*operationsContractOperation{
		"get": path.Get, "post": path.Post, "put": path.Put, "delete": path.Delete,
	}
}

type expectedOperation struct{ route, capability string }

func TestPlatformOperationsRoutesMatchOpenAPIContract(t *testing.T) {
	document, err := os.ReadFile("../../../../contracts/http/platform-operations-v1.paths.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract operationsPathContract
	if err = json.Unmarshal(document, &contract); err != nil {
		t.Fatal(err)
	}
	expected := map[string]map[string]expectedOperation{
		"health":        {"get": {"GET /api/platform/operations/health", "platform.operations.health.read"}},
		"traces":        {"get": {"GET /api/platform/operations/traces", "platform.operations.traces.read"}},
		"traceSearch":   {"post": {"POST /api/platform/operations/traces/search", "platform.operations.traces.read"}},
		"traceStats":    {"get": {"GET /api/platform/operations/traces/stats", "platform.operations.traces.read"}},
		"traceSpan":     {"get": {"GET /api/platform/operations/traces/spans/{span_id}", "platform.operations.traces.read"}},
		"trace":         {"get": {"GET /api/platform/operations/traces/{correlation_id}", "platform.operations.traces.read"}},
		"traceTimeline": {"get": {"GET /api/platform/operations/traces/{correlation_id}/timeline", "platform.operations.traces.read"}},
		"rules": {
			"get":  {"GET /api/platform/operations/alert-rules", "platform.operations.alerts.read"},
			"post": {"POST /api/platform/operations/alert-rules", "platform.operations.alerts.manage"},
		},
		"rule": {
			"get":    {"GET /api/platform/operations/alert-rules/{alert_rule_id}", "platform.operations.alerts.read"},
			"delete": {"DELETE /api/platform/operations/alert-rules/{alert_rule_id}", "platform.operations.alerts.manage"},
			"put":    {"PUT /api/platform/operations/alert-rules/{alert_rule_id}", "platform.operations.alerts.manage"},
		},
		"ruleTest": {"post": {"POST /api/platform/operations/alert-rules/{alert_rule_id}/test", "platform.operations.alerts.manage"}},
		"events":   {"get": {"GET /api/platform/operations/alert-events", "platform.operations.alerts.read"}},
		"event":    {"get": {"GET /api/platform/operations/alert-events/{alert_event_id}", "platform.operations.alerts.read"}},
		"permissionCacheFlush": {
			"post": {"POST /api/platform/operations/permission-cache/flush", "platform.operations.permission_cache.flush"},
		},
	}
	actualRoutes := make(map[string]struct{})
	handler, _, _ := httpHandlerFixture()
	for _, route := range handler.Routes() {
		if strings.HasPrefix(route.Pattern, "/api/platform/operations/") {
			actualRoutes[route.Method+" "+route.Pattern] = struct{}{}
		}
	}
	var contracted []string
	for key, requirements := range expected {
		item, found := contract.PathItems[key]
		if !found {
			t.Fatalf("missing path item %q", key)
		}
		for method, operation := range item.operations() {
			if operation == nil {
				continue
			}
			requirement, found := requirements[method]
			if !found || operation.Capability != requirement.capability {
				t.Fatalf("%s %s capability = %q", key, method, operation.Capability)
			}
			if _, found = actualRoutes[requirement.route]; !found || operation.OperationID == "" {
				t.Fatalf("contracted route is not reachable: %s", requirement.route)
			}
			contracted = append(contracted, requirement.route)
		}
	}
	if len(contracted) != len(actualRoutes) {
		t.Fatalf("contract routes = %d, runtime routes = %d", len(contracted), len(actualRoutes))
	}
}
