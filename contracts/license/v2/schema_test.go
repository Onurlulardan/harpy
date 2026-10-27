package licensev2

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestLicenseSchemasAcceptExactContracts(t *testing.T) {
	tests := []struct {
		path  string
		value map[string]any
	}{
		{"license-snapshot.schema.json", validSnapshotContract()},
		{"license-issue-request.schema.json", validIssueContract()},
		{"license-trust-bundle.schema.json", validTrustContract()},
		{"license-derived-offering.schema.json", validDerivedOfferingContract()},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			if err := compileLicenseSchema(t, test.path).Validate(test.value); err != nil {
				t.Fatalf("valid license contract rejected: %v", err)
			}
		})
	}
}

func TestLicenseSchemasRejectAuthorityDrift(t *testing.T) {
	unknown := validSnapshotContract()
	unknown["temporary"] = true
	wrongMode := validSnapshotContract()
	wrongMode["payload"].(map[string]any)["deployment_mode"] = "saas"
	weakIssue := validIssueContract()
	weakIssue["current_sequence"] = 2
	rootReuse := validTrustContract()
	rootReuse["purpose"] = "protected-authority"
	weakOffering := validDerivedOfferingContract()
	weakOffering["booleans"].([]any)[0].(map[string]any)["temporary"] = true
	tests := []struct {
		path  string
		value map[string]any
	}{
		{"license-snapshot.schema.json", unknown},
		{"license-snapshot.schema.json", wrongMode},
		{"license-issue-request.schema.json", weakIssue},
		{"license-trust-bundle.schema.json", rootReuse},
		{"license-derived-offering.schema.json", weakOffering},
	}
	for _, test := range tests {
		if err := compileLicenseSchema(t, test.path).Validate(test.value); err == nil {
			t.Fatalf("unsafe license contract accepted by %s", test.path)
		}
	}
}

func compileLicenseSchema(t *testing.T, path string) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	entries, err := filepath.Glob("*.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	const base = "https://harpy.io/contracts/license/v2/"
	for _, entry := range entries {
		document, err := os.ReadFile(entry)
		if err != nil {
			t.Fatal(err)
		}
		var value any
		if err = json.Unmarshal(document, &value); err != nil {
			t.Fatal(err)
		}
		if err = compiler.AddResource(base+entry, value); err != nil {
			t.Fatal(err)
		}
	}
	schema, err := compiler.Compile(base + path)
	if err != nil {
		t.Fatal(err)
	}
	return schema
}
