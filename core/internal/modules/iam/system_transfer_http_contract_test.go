package iam

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type systemTransferPathsFragment struct {
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

type systemTransferContractOperation struct {
	Permission string `json:"x-harpy-permission-template"`
	Authority  string `json:"x-harpy-authority"`
}

func TestSystemTransferHTTPRoutesMatchADR0103(t *testing.T) {
	expected := map[string][2]string{
		"POST " + systemExportCollection:         {"iam.{resource}.export", SystemExportSubmitOperation},
		"GET " + systemExportMember:              {"iam.{resource}.export", SystemExportGetOperation},
		"DELETE " + systemExportMember:           {"iam.{resource}.export", SystemExportCancelOperation},
		"GET " + systemExportMember + "/content": {"iam.{resource}.export", SystemExportDownloadOperation},
		"POST " + systemImportSources:            {"iam.{resource}.import", SystemImportSourceOperation},
		"POST " + systemImportCollection:         {"iam.{resource}.import", SystemImportSubmitOperation},
		"GET " + systemImportMember:              {"iam.{resource}.import", SystemImportGetOperation},
		"DELETE " + systemImportMember:           {"iam.{resource}.import", SystemImportCancelOperation},
		"GET " + systemImportMember + "/report":  {"iam.{resource}.import", SystemImportReportOperation},
	}
	document, err := os.ReadFile("../../../../contracts/http/iam-system-transfer-v1.paths.json")
	if err != nil {
		t.Fatal(err)
	}
	var fragment systemTransferPathsFragment
	if json.Unmarshal(document, &fragment) != nil {
		t.Fatal("IAM transfer path fragment is invalid")
	}
	routes := new(SystemTransferHTTP).Routes()
	if len(routes) != len(expected)+30 {
		t.Fatalf("routes=%d want=%d", len(routes), len(expected)+30)
	}
	for _, route := range routes[:len(expected)] {
		want, found := expected[route.Method+" "+route.Pattern]
		if !found {
			t.Fatalf("unexpected route %s %s", route.Method, route.Pattern)
		}
		var operation systemTransferContractOperation
		methods := fragment.Paths[route.Pattern]
		if json.Unmarshal(methods[strings.ToLower(route.Method)], &operation) != nil ||
			operation.Permission != want[0] || operation.Authority != want[1] {
			t.Fatalf("contract %s %s = %+v", route.Method, route.Pattern, operation)
		}
	}
}

func TestSystemTransferSchemaIsRedactedAndInviteOnly(t *testing.T) {
	document, err := os.ReadFile("../../../../contracts/http/iam-system-transfer-v1.schemas.json")
	if err != nil || !json.Valid(document) {
		t.Fatalf("schema read error=%v", err)
	}
	source := string(document)
	for _, marker := range []string{`"duplicate_action"`, `"skip"`, `"update"`,
		`"tabular-mapping.v2"`, `"invitation"`, `"source_digest"`, `"result_digest"`} {
		if !strings.Contains(source, marker) {
			t.Fatalf("schema lacks %s", marker)
		}
	}
	for _, forbidden := range []string{`"password"`, `"email"`, `"mapped_values"`,
		`"source_values"`, `"session"`, `"token"`} {
		if strings.Contains(strings.ToLower(source), forbidden) {
			t.Fatalf("schema exposes %s", forbidden)
		}
	}
}
