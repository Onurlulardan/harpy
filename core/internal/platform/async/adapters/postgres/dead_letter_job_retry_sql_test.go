package postgres

import (
	"strings"
	"testing"
)

func TestDeadLetterGapJobRetryLocksCanonicalEvidence(t *testing.T) {
	required := []string{
		"FOR UPDATE OF source",
		"FOR UPDATE OF event",
		"gap_root.full_payload_expires_at > retry_clock.observed_at",
		"gap_root.payload_canonical IS NOT NULL",
		"event.payload_canonical IS NOT NULL",
		"event.delivery_state = 'delivered'",
		"'producer_authority'->>'group_generation'",
		"'producer_authority'->>'cutover_generation'",
		"platform.platform_async_assert_current_authority_v1(",
		"AS authority_assertion",
		"source.kind =",
		"'platform_async.event_gap_reconcile.' || source.authority_domain",
		"source.id AS source_job_id",
	}
	for _, marker := range required {
		if !strings.Contains(insertDeadLetterJobRetrySQL, marker) {
			t.Fatalf("dead-letter gap retry SQL is missing %q", marker)
		}
	}
}

func TestDeadLetterNonGapJobRetryKeepsGenericBranch(t *testing.T) {
	required := []string{
		"eligible_source AS MATERIALIZED",
		"UNION ALL",
		"<> 'platform_async.event_gap_reconcile.'",
		"source.effect_idempotency_key",
		"source.max_attempts",
		"source.id",
	}
	for _, marker := range required {
		if !strings.Contains(insertDeadLetterJobRetrySQL, marker) {
			t.Fatalf("generic dead-letter retry SQL is missing %q", marker)
		}
	}
}
