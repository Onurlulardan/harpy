package main

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

type policyDocument struct {
	Statement []struct {
		Resource []string `json:"Resource"`
	} `json:"Statement"`
}

func TestS3RuntimePolicyNamesOnlyRuntimeBuckets(t *testing.T) {
	content, err := os.ReadFile("../../scripts/s3-runtime-policy.json")
	if err != nil {
		t.Fatal(err)
	}
	var policy policyDocument
	if err := json.Unmarshal(content, &policy); err != nil {
		t.Fatal(err)
	}
	actual := make([]string, 0, 12)
	for _, statement := range policy.Statement {
		actual = append(actual, statement.Resource...)
	}
	sort.Strings(actual)
	buckets := []string{
		"harpy-documents", "harpy-documents-legacy",
		"harpy-analytics-dashboard-exports",
		"harpy-form-exports", "harpy-form-import-sources",
		"harpy-iam-system-exports", "harpy-iam-system-import-sources",
	}
	expected := make([]string, 0, 12)
	for _, bucket := range buckets {
		expected = append(expected, "arn:aws:s3:::"+bucket, "arn:aws:s3:::"+bucket+"/*")
	}
	sort.Strings(expected)
	if len(actual) != len(expected) {
		t.Fatalf("policy has %d resources, want %d", len(actual), len(expected))
	}
	for index := range expected {
		if actual[index] != expected[index] {
			t.Fatalf("resource %d = %q, want %q", index, actual[index], expected[index])
		}
	}
}
