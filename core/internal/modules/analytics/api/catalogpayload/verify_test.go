package catalogpayload

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
)

func TestVerifyAcceptsStructurallyBoundFirstPartyAnalytics(t *testing.T) {
	componentID := "018f0000-0002-7000-8000-000000000009"
	definition := canonicalAnalyticsJSON(t, map[string]any{
		"dataSources": []any{map[string]any{"formSlug": "project-task"}},
		"fields":      []any{map[string]any{"field": "id"}},
	})
	binding := canonicalAnalyticsJSON(t, map[string]any{
		"slug": "project-task-count", "name": "Task count",
		"description": "Project tasks", "definition": json.RawMessage(definition),
	})
	legacyHash, err := canonicaljsonv1.DigestHex(binding, MaximumPayloadBytes)
	if err != nil {
		t.Fatal(err)
	}
	payload := Payload{
		ContractVersion: catalogpackagev1.ContractAnalytics,
		ComponentID:     componentID, AnalyticsID: componentID, Kind: "endpoint",
		Slug: "project-task-count", Name: "Task count",
		Description: "Project tasks", LegacyHash: legacyHash,
		SourceContract: "harpy-v1.enterprise-seed.project",
		Definition:     definition,
	}
	document := canonicalAnalyticsJSON(t, payload)
	digest := sha256.Sum256(document)
	verified, err := Verify(document, componentID, hex.EncodeToString(digest[:]))
	if err != nil || verified.Value().Kind != "endpoint" {
		t.Fatalf("verify project analytics: value=%#v err=%v", verified.Value(), err)
	}
	payload.SourceContract = "harpy-v1.enterprise-seed.Project"
	mismatched := canonicalAnalyticsJSON(t, payload)
	mismatchedDigest := sha256.Sum256(mismatched)
	if _, err = Verify(
		mismatched, componentID, hex.EncodeToString(mismatchedDigest[:]),
	); err == nil {
		t.Fatal("analytics accepted a non-canonical source contract")
	}
}

func TestVerifyEnforcesDatabaseTextByteLimits(t *testing.T) {
	componentID := "018f0000-0002-7000-8000-000000000009"
	for _, test := range []struct {
		name        string
		displayName string
		description string
		wantError   bool
	}{
		{name: "maximum name", displayName: strings.Repeat("n", 100), description: "ok"},
		{name: "oversized name", displayName: strings.Repeat("n", 101), description: "ok", wantError: true},
		{name: "maximum description", displayName: "ok", description: strings.Repeat("d", 2048)},
		{name: "oversized description", displayName: "ok", description: strings.Repeat("d", 2049), wantError: true},
		{name: "multibyte name uses bytes", displayName: strings.Repeat("ş", 51), description: "ok", wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			definition := canonicalAnalyticsJSON(t, map[string]any{
				"dataSources": []any{map[string]any{"formSlug": "project-task"}},
				"fields":      []any{map[string]any{"field": "id"}},
			})
			binding := canonicalAnalyticsJSON(t, map[string]any{
				"slug": "project-task-count", "name": test.displayName,
				"description": test.description, "definition": json.RawMessage(definition),
			})
			legacyHash, err := canonicaljsonv1.DigestHex(binding, MaximumPayloadBytes)
			if err != nil {
				t.Fatal(err)
			}
			document := canonicalAnalyticsJSON(t, Payload{
				ContractVersion: catalogpackagev1.ContractAnalytics,
				ComponentID:     componentID, AnalyticsID: componentID, Kind: "endpoint",
				Slug: "project-task-count", Name: test.displayName,
				Description: test.description, LegacyHash: legacyHash,
				SourceContract: "harpy-v1.enterprise-seed.project",
				Definition:     definition,
			})
			digest := sha256.Sum256(document)
			_, err = Verify(document, componentID, hex.EncodeToString(digest[:]))
			if (err != nil) != test.wantError {
				t.Fatalf("Verify() error = %v, wantError %v", err, test.wantError)
			}
		})
	}
}

func canonicalAnalyticsJSON(t *testing.T, value any) []byte {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumPayloadBytes)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}
