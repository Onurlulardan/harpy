package main

import (
	"encoding/json"
	"testing"
	"time"
)

type generatedIAMActive struct {
	ActivatedAt string `json:"activated_at"`
	SignUntil   string `json:"sign_until"`
	IssueUntil  string `json:"issue_until"`
}

func TestGeneratedIAMKeyringsHaveExactActiveLifecycles(t *testing.T) {
	set, now := generatedSetForTest(t)
	var access struct {
		Active generatedIAMActive `json:"active"`
	}
	if err := json.Unmarshal(set.files["iam-access-token-keyring.json"], &access); err != nil {
		t.Fatal(err)
	}
	activated := mustTime(t, access.Active.ActivatedAt)
	signUntil := mustTime(t, access.Active.SignUntil)
	if signUntil.Sub(activated) != 30*24*time.Hour || now.Before(activated) ||
		!now.Before(signUntil) {
		t.Fatal("access-token lifecycle differs from the runtime contract")
	}
	for _, target := range iamHMACTargets {
		var document struct {
			Purpose string             `json:"purpose"`
			Active  generatedIAMActive `json:"active"`
		}
		if err := json.Unmarshal(set.files[target.file], &document); err != nil {
			t.Fatal(err)
		}
		activeAt := mustTime(t, document.Active.ActivatedAt)
		issueUntil := mustTime(t, document.Active.IssueUntil)
		if document.Purpose != target.purpose ||
			issueUntil.Sub(activeAt) != 90*24*time.Hour ||
			now.Before(activeAt) || !now.Before(issueUntil) {
			t.Fatalf("invalid HMAC lifecycle for %s", target.purpose)
		}
	}
}

func mustTime(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func TestGeneratedArtifactKEKsAreCryptographicallySeparated(t *testing.T) {
	set, _ := generatedSetForTest(t)
	seen := make(map[string]string)
	for _, target := range artifactKEKTargets {
		var document struct {
			Current struct {
				Key string `json:"key"`
			} `json:"current"`
		}
		if err := json.Unmarshal(set.files[target.file], &document); err != nil {
			t.Fatal(err)
		}
		if previous := seen[document.Current.Key]; previous != "" {
			t.Fatalf("%s reuses key material from %s", target.file, previous)
		}
		seen[document.Current.Key] = target.file
	}
}
