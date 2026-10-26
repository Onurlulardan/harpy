//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"hash"
	"testing"

	"github.com/google/uuid"
)

const (
	analyticsCutoverQuery     = "018f2000-0000-7000-8000-000000000170"
	analyticsCutoverPublished = "018f2000-0000-7000-8000-000000000171"
	analyticsCutoverDraft     = "018f2000-0000-7000-8000-000000000172"
	analyticsCutoverRelease   = "018f2000-0000-7000-8000-000000000173"
	analyticsCutoverRequester = "018f2000-0000-7000-8000-000000000174"
	analyticsCutoverApprover  = "018f2000-0000-7000-8000-000000000175"
	analyticsCutoverResource  = "018f2000-0000-7000-8000-000000000176"
	analyticsCutoverVersion   = "018f2000-0000-7000-8000-000000000177"
)

type analyticsCutoverValue struct {
	sourceDocument, targetDocument string
	sourceSHA256, targetSHA256     string
}

func loadAnalyticsCutoverValue(t *testing.T, ctx context.Context,
	database *sql.DB,
) analyticsCutoverValue {
	t.Helper()
	var source string
	if err := database.QueryRowContext(ctx, `SELECT to_jsonb(endpoint)::text
FROM t_analytics_a.openapi_endpoints AS endpoint WHERE endpoint.id=$1`,
		analyticsCutoverEndpoint).Scan(&source); err != nil {
		t.Fatal(err)
	}
	sourceDigest := sha256.Sum256([]byte(source))
	manifest := analyticsCutoverDependencyManifest()
	request := hexRepeat("55")
	requestBytes, _ := hex.DecodeString(request)
	manifestBytes, _ := hex.DecodeString(manifest)
	approvalDigest := sha256.Sum256(append(requestBytes, manifestBytes...))
	dependency := map[string]any{
		"ordinal": 0, "owner_domain": "iam", "resource_kind": "role",
		"resource_id": analyticsCutoverResource, "resource_version_id": analyticsCutoverVersion,
		"owner_generation": 1, "catalog_version": 1, "query_contract_version": 1,
		"policy_version": 1, "data_revision_capability": "revision-2",
		"dependency_sha256": hexRepeat("22"), "created_at": "2026-07-21T10:00:00.123Z"}
	version := func(id string, number int, state string) map[string]any {
		publishedBy, publishedAt := any(nil), any(nil)
		if state == "published" {
			publishedBy, publishedAt = analyticsCutoverRequester, "2026-07-21T10:02:00.123Z"
		}
		return map[string]any{"id": id, "query_id": analyticsCutoverQuery,
			"version_no": number, "version_state": state, "row_revision": 1,
			"name": "Migrated report", "query_document": map[string]any{
				"method": "GET", "path": "/legacy/report"},
			"parameter_document": map[string]any{}, "output_document": map[string]any{},
			"manifest_sha256": manifest, "cache_policy": "revision_bound",
			"created_by": analyticsCutoverRequester, "updated_by": analyticsCutoverRequester,
			"created_at":   "2026-07-21T10:00:00.123Z",
			"updated_at":   "2026-07-21T10:02:00.123Z",
			"published_by": publishedBy, "published_at": publishedAt}
	}
	differential := analyticsCutoverDifferential(sourceDigest[:])
	target := map[string]any{
		"query": map[string]any{"id": analyticsCutoverQuery, "lifecycle_state": "active",
			"row_revision": 1, "reference_epoch": 1,
			"current_draft_version_id": analyticsCutoverDraft,
			"created_by":               analyticsCutoverRequester, "updated_by": analyticsCutoverRequester,
			"created_at": "2026-07-21T10:00:00.123Z",
			"updated_at": "2026-07-21T10:02:00.123Z"},
		"published_version": version(analyticsCutoverPublished, 1, "published"),
		"draft_version":     version(analyticsCutoverDraft, 2, "draft"),
		"dependencies":      []any{dependency}, "differential": differential,
		"release": map[string]any{"id": analyticsCutoverRelease, "resource_kind": "query",
			"resource_version_id": analyticsCutoverPublished, "audience": "public",
			"lifecycle_state": "active", "cache_policy": "revision_bound",
			"manifest_sha256": manifest, "requester_id": analyticsCutoverRequester,
			"approver_id": analyticsCutoverApprover, "row_revision": 4, "generation": 1,
			"public_id": hexRepeat16("33"), "capability_expires_at": nil,
			"created_at":   "2026-07-21T10:00:00.123Z",
			"updated_at":   "2026-07-21T10:04:00.123Z",
			"submitted_at": "2026-07-21T10:01:00.123Z",
			"approved_at":  "2026-07-21T10:03:00.123Z",
			"activated_at": "2026-07-21T10:04:00.123Z", "revoked_at": nil,
			"requester_step_up_at": "2026-07-21T10:00:30.123Z"},
		"approval": map[string]any{"release_id": analyticsCutoverRelease,
			"requester_id":         analyticsCutoverRequester,
			"requester_step_up_at": "2026-07-21T10:00:30.123Z",
			"approver_id":          analyticsCutoverApprover,
			"approver_step_up_at":  "2026-07-21T10:02:30.123Z",
			"approved_at":          "2026-07-21T10:03:00.123Z", "request_sha256": request,
			"evidence_sha256": hex.EncodeToString(approvalDigest[:])},
		"public_pointer": map[string]any{"public_id": hexRepeat16("33"),
			"tenant_id": analyticsTenantA, "release_id": analyticsCutoverRelease,
			"resource_kind": "query", "resource_version_id": analyticsCutoverPublished,
			"audience": "public", "state": "active", "generation": 1,
			"capability_sha256": nil, "capability_expires_at": nil,
			"activated_at": "2026-07-21T10:04:00.123Z", "revoked_at": nil},
		"legacy_route": map[string]any{"legacy_tenant_identity": analyticsTenantA,
			"method": "GET", "path": "/legacy/report", "slug": "legacy-report",
			"tenant_id": analyticsTenantA, "legacy_endpoint_id": analyticsCutoverEndpoint,
			"public_id": hexRepeat16("33"), "release_id": analyticsCutoverRelease,
			"state": "active", "hit_count": 0, "last_hit_at": nil,
			"consumers_acked_at":         "2026-07-21T10:05:00.123Z",
			"remove_after":               "2026-08-21T10:05:00.123Z",
			"consumer_inventory_sha256":  hexRepeat("66"),
			"retirement_evidence_sha256": nil, "retirement_observed_at": nil,
			"retired_at": nil},
	}
	targetText := analyticsCanonicalJSON(t, ctx, database, target)
	targetDigest := sha256.Sum256([]byte(targetText))
	return analyticsCutoverValue{source, targetText, hex.EncodeToString(sourceDigest[:]),
		hex.EncodeToString(targetDigest[:])}
}

func analyticsCutoverDependencyManifest() string {
	digest := sha256.New()
	writeAnalyticsFrame(digest, []byte("iam"))
	writeAnalyticsFrame(digest, []byte("role"))
	writeAnalyticsFrame(digest, analyticsUUIDBytes(analyticsCutoverResource))
	writeAnalyticsFrame(digest, analyticsUUIDBytes(analyticsCutoverVersion))
	writeAnalyticsFrame(digest, []byte("revision-2"))
	dependency, _ := hex.DecodeString(hexRepeat("22"))
	writeAnalyticsFrame(digest, dependency)
	for range 3 {
		_ = binary.Write(digest, binary.BigEndian, uint64(1))
	}
	return hex.EncodeToString(digest.Sum(nil))
}

func writeAnalyticsFrame(digest hash.Hash, value []byte) {
	_ = binary.Write(digest, binary.BigEndian, uint32(len(value)))
	_, _ = digest.Write(value)
}

func analyticsCutoverDifferential(source []byte) map[string]any {
	result, wire, image := hexRepeat("77"), hexRepeat("88"), hexRepeat("99")
	digest := sha256.New()
	_, _ = digest.Write([]byte("analytics-v1-adoption-differential.v1"))
	for _, value := range [][]byte{analyticsUUIDBytes(analyticsCutoverEndpoint),
		mustHex(result), mustHex(result), mustHex(wire), mustHex(image), source} {
		_, _ = digest.Write([]byte{0})
		_, _ = digest.Write(value)
	}
	return map[string]any{"source_result_sha256": result, "target_result_sha256": result,
		"wire_contract_sha256": wire, "compatible_image_sha256": image,
		"reverse_source_sha256": hex.EncodeToString(source),
		"evidence_sha256":       hex.EncodeToString(digest.Sum(nil))}
}

func analyticsCanonicalJSON(t *testing.T, ctx context.Context,
	database *sql.DB, value any,
) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var result string
	if err = database.QueryRowContext(ctx, `SELECT $1::jsonb::text`, string(encoded)).Scan(&result); err != nil {
		t.Fatal(err)
	}
	return result
}

func hexRepeat(value string) string   { return stringRepeat(value, 32) }
func hexRepeat16(value string) string { return stringRepeat(value, 16) }
func stringRepeat(value string, count int) string {
	result := ""
	for range count {
		result += value
	}
	return result
}
func mustHex(value string) []byte { decoded, _ := hex.DecodeString(value); return decoded }
func analyticsUUIDBytes(value string) []byte {
	identifier := uuid.MustParse(value)
	return identifier[:]
}
