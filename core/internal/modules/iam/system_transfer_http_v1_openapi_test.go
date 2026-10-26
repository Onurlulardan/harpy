package iam

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type systemTransferV1Contract struct {
	Paths map[string]map[string]struct {
		Deprecated bool              `json:"deprecated"`
		Successor  string            `json:"x-harpy-successor"`
		Responses  map[string]any    `json:"responses"`
		Security   []json.RawMessage `json:"security"`
	} `json:"paths"`
}

func TestSystemTransferV1RoutesMatchVersionedOpenAPI(t *testing.T) {
	document, err := os.ReadFile(
		"../../../../contracts/http/iam-system-transfer-v1-compat.paths.json")
	if err != nil {
		t.Fatal(err)
	}
	var contract systemTransferV1Contract
	if json.Unmarshal(document, &contract) != nil || len(contract.Paths) != 30 {
		t.Fatalf("invalid V1 compatibility fragment paths=%d", len(contract.Paths))
	}
	routes := new(SystemTransferHTTP).Routes()[9:]
	for _, route := range routes {
		operation, found := contract.Paths[route.Pattern][strings.ToLower(route.Method)]
		if !found || !operation.Deprecated || operation.Successor == "" ||
			len(operation.Security) != 0 || operation.Responses["426"] == nil {
			t.Fatalf("contract missing retired route %s %s", route.Method, route.Pattern)
		}
	}
	schemas, err := os.ReadFile(
		"../../../../contracts/http/iam-system-transfer-v1-compat.schemas.json")
	if err != nil || !json.Valid(schemas) {
		t.Fatalf("invalid V1 compatibility schemas: %v", err)
	}
	for _, marker := range []string{"Cache-Control", "Deprecation", "Link",
		"successor-version", "ErrorEnvelope"} {
		if !strings.Contains(string(schemas), marker) {
			t.Fatalf("V1 compatibility response lacks %s", marker)
		}
	}
}
