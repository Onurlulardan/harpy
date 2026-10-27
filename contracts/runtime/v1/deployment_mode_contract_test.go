package runtimev1

import "testing"

func TestAPIDeploymentModeIsOptionalAndExact(t *testing.T) {
	schema := compileSchema(t, "core-api-config.schema.json")
	missing := apiOnPremContractSnapshot()
	if err := schema.Validate(missing); err != nil {
		t.Fatalf("missing deployment mode must select on-premise: %v", err)
	}
	for _, invalid := range []string{"", " ", "SAAS", "hosted"} {
		snapshot := profileSnapshot("api")
		snapshot["variables"].(map[string]any)["DEPLOYMENT_MODE"] = invalid
		if err := schema.Validate(snapshot); err == nil {
			t.Fatalf("invalid deployment mode %q was accepted", invalid)
		}
	}
}

func TestWorkerDeploymentModeIsOptionalAndExact(t *testing.T) {
	schema := compileSchema(t, "core-worker-config.schema.json")
	missing := profileSnapshot("worker")
	if err := schema.Validate(missing); err != nil {
		t.Fatalf("missing worker deployment mode must select on-premise: %v", err)
	}
	for _, value := range []string{"onpremise", "saas"} {
		snapshot := profileSnapshot("worker")
		snapshot["variables"].(map[string]any)["DEPLOYMENT_MODE"] = value
		if value == "saas" {
			selectWorkerSaaSContractSnapshot(snapshot["variables"].(map[string]any))
		}
		if err := schema.Validate(snapshot); err != nil {
			t.Fatalf("worker deployment mode %q rejected: %v", value, err)
		}
	}
}
