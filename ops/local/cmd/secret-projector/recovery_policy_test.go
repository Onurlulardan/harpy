package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type recoveryPolicy struct {
	Statement []struct {
		Action   []string `json:"Action"`
		Resource []string `json:"Resource"`
	} `json:"Statement"`
}

func TestRecoveryPoliciesAreBucketAndPurposeScoped(t *testing.T) {
	tests := []struct {
		name, bucket string
		required     []string
		forbidden    []string
	}{
		{"recovery-write-policy.json", "harpy-recovery",
			[]string{"s3:PutObject", "s3:GetObject"}, []string{"s3:DeleteObject"}},
		{"recovery-retention-delete-policy.json", "harpy-recovery",
			[]string{"s3:DeleteObject", "s3:DeleteObjectVersion"},
			[]string{"s3:PutObject", "s3:GetObject", "s3:GetObjectVersion"}},
		{"recovery-restore-read-policy.json", "harpy-recovery",
			[]string{"s3:GetObject", "s3:GetObjectVersion"},
			[]string{"s3:PutObject", "s3:DeleteObject"}},
		{"recovery-documents-source-read-policy.json", "harpy-documents",
			[]string{"s3:GetObject", "s3:GetObjectVersion"},
			[]string{"s3:PutObject", "s3:DeleteObject"}},
		{"recovery-documents-rehydrate-write-policy.json", "harpy-documents-rehydrate",
			[]string{"s3:PutObject", "s3:GetObject"}, []string{"s3:DeleteObject"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content, err := os.ReadFile("../../scripts/" + test.name)
			if err != nil {
				t.Fatal(err)
			}
			var policy recoveryPolicy
			if json.Unmarshal(content, &policy) != nil || len(policy.Statement) == 0 {
				t.Fatal("policy is not valid JSON")
			}
			actions := map[string]bool{}
			for _, statement := range policy.Statement {
				for _, action := range statement.Action {
					actions[action] = true
				}
				for _, resource := range statement.Resource {
					prefix := "arn:aws:s3:::" + test.bucket
					if resource != prefix && resource != prefix+"/*" {
						t.Fatalf("foreign resource %s", resource)
					}
				}
			}
			for _, action := range test.required {
				if !actions[action] {
					t.Fatalf("required action %s is missing", action)
				}
			}
			for _, action := range test.forbidden {
				if actions[action] {
					t.Fatalf("forbidden action %s is present", action)
				}
			}
		})
	}
}

func TestRecoveryBucketIsCreatedWithObjectLockOnce(t *testing.T) {
	content, err := os.ReadFile("../../scripts/init-recovery-s3.sh")
	if err != nil {
		t.Fatal(err)
	}
	script := string(content)
	required := []string{
		"if ! mc stat recovery/harpy-recovery",
		"mc mb --with-lock recovery/harpy-recovery",
		"mc version enable recovery/harpy-recovery",
		"mc anonymous set none recovery/harpy-recovery",
		"mc retention set --default COMPLIANCE 30d recovery/harpy-recovery",
	}
	for _, value := range required {
		if !strings.Contains(script, value) {
			t.Fatalf("recovery init is missing %q", value)
		}
	}
	if strings.Contains(script, "mc mb --ignore-existing recovery/harpy-recovery") {
		t.Fatal("recovery bucket creation could bypass Object Lock")
	}
}
