package licensev2

import "strings"

const fixtureInstant = "2026-07-22T12:00:00Z"

func validSnapshotContract() map[string]any {
	digest, signature := strings.Repeat("a", 64), strings.Repeat("A", 86)
	return map[string]any{
		"schema": SnapshotEnvelopeSchema, "kid": "license-1",
		"algorithm": SignatureAlgorithm, "canonicalization": Canonicalization,
		"payload_sha256": digest, "signature": signature,
		"payload": map[string]any{
			"schema": SnapshotSchema, "issuer": "harpy",
			"audience": SnapshotAudience, "product": SnapshotProduct,
			"license_id": fixtureUUID(1), "licensee_organization_id": fixtureUUID(2),
			"environment_id": fixtureUUID(3), "deployment_mode": "onpremise",
			"delivery_kind": "connected", "sequence": 1, "status": "active",
			"issued_at": fixtureInstant, "not_before": fixtureInstant,
			"refresh_after": fixtureInstant, "grant_valid_until": fixtureInstant,
			"grace_until": fixtureInstant, "commercial_expires_at": fixtureInstant,
			"release_channel": "stable",
			"features": map[string]any{
				"formBuilder": true, "workflowBuilder": true, "apiBuilder": true,
				"dashboardBuilder": true, "aiAssistant": true,
			},
			"limits": map[string]any{
				"maxUsers": 0, "maxForms": 0, "maxWorkflows": 0,
				"maxApiEndpoints": 0, "maxDashboards": 0, "maxTenants": 1,
			},
		},
	}
}

func validIssueContract() map[string]any {
	return map[string]any{
		"schema": IssueRequestSchema, "operation_id": fixtureUUID(4),
		"license_id": fixtureUUID(1), "licensee_organization_id": fixtureUUID(2),
		"environment_id": fixtureUUID(3), "deployment_mode": "onpremise",
		"delivery_kind": "connected", "nonce": strings.Repeat("A", 43),
		"current_sequence": 0, "current_digest": nil,
		"component_versions": []any{map[string]any{
			"component": "core-worker", "version": "v2.0.0",
		}},
	}
}

func validTrustContract() map[string]any {
	key := func(id string) map[string]any {
		return map[string]any{
			"kid": id, "algorithm": SignatureAlgorithm,
			"public_key": strings.Repeat("A", 43), "status": "active",
			"not_before": fixtureInstant, "sign_until": fixtureInstant,
			"verify_until": fixtureInstant,
		}
	}
	return map[string]any{
		"schema": TrustBundleSchema, "purpose": SnapshotEnvelopeSchema,
		"roots": []any{key("release-1")},
		"keysets": []any{map[string]any{
			"schema": KeysetEnvelopeSchema, "kid": "release-1",
			"algorithm": SignatureAlgorithm, "canonicalization": Canonicalization,
			"payload_sha256": strings.Repeat("a", 64),
			"signature":      strings.Repeat("A", 86),
			"payload": map[string]any{
				"schema": KeysetSchema, "issuer": "harpy", "audience": KeysetAudience,
				"sequence": 1, "issued_at": fixtureInstant, "not_before": fixtureInstant,
				"expires_at": fixtureInstant, "snapshot_signers": []any{key("license-1")},
			},
		}},
	}
}

func fixtureUUID(last int) string {
	return "00000000-0000-7000-8000-00000000000" + string(rune('0'+last))
}

func validDerivedOfferingContract() map[string]any {
	booleanKeys := []string{
		"analytics.dashboard_builder.enabled", "api.access",
		"api.builder.enabled", "forms.builder.enabled", "workflow.builder.enabled",
		"workflow.enabled",
	}
	limitKeys := []string{
		"analytics.active_dashboards", "api.active_endpoints", "documents.storage_bytes",
		"forms.active_forms", "forms.submissions", "iam.active_users",
		"workflow.active_definitions",
	}
	booleans := make([]any, len(booleanKeys))
	limits := make([]any, len(limitKeys))
	for index, key := range booleanKeys {
		booleans[index] = map[string]any{"key": key, "value": true}
	}
	for index, key := range limitKeys {
		limits[index] = map[string]any{"key": key, "value": 0, "unlimited": true}
	}
	return map[string]any{
		"schema": DerivedOfferingSchema, "offering_version_id": fixtureUUID(4),
		"revision": 1, "effective_at": fixtureInstant,
		"license_id": fixtureUUID(1), "licensee_organization_id": fixtureUUID(2),
		"environment_id": fixtureUUID(3), "issuer": "harpy",
		"snapshot_sequence": 1, "snapshot_sha256": strings.Repeat("a", 64),
		"release_channel": "stable", "booleans": booleans, "limits": limits,
	}
}
