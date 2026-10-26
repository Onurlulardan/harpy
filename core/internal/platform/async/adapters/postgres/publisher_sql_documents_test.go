package postgres

import (
	"strings"
	"testing"
)

func TestPublisherSQLDocumentsCarryAuthorityOrderingAndLeaseFences(t *testing.T) {
	t.Parallel()
	claimMarkers := []string{
		"FOR UPDATE OF event SKIP LOCKED",
		"event.authority_group_generation = $4",
		"event.cutover_generation = $6",
		"predecessor.delivery_state = 'delivered'",
		"prior.delivery_state <> 'delivered'",
		"LIMIT $8",
		"interval '60 seconds'",
	}
	for _, marker := range claimMarkers {
		if !strings.Contains(claimOutboxSQL, marker) {
			t.Fatalf("claim SQL lacks %q", marker)
		}
	}
	for name, document := range map[string]string{
		"heartbeat": heartbeatPublishSQL,
		"confirm":   confirmPublishSQL,
		"release":   releasePublishSQL,
	} {
		for _, marker := range []string{
			"event.lease_owner_id = $8", "event.lease_generation = $9",
			"event.lease_until > publish_clock.observed_at",
		} {
			if !strings.Contains(document, marker) {
				t.Fatalf("%s SQL lacks %q", name, marker)
			}
		}
	}
}

func TestPublisherSQLDocumentsUseCurrentV2AuthorityAndStableRecovery(t *testing.T) {
	t.Parallel()
	for _, marker := range []string{
		"group_row.state = 'v2'", "group_row.writer = 'v2'",
		"domain_row.state = 'v2'", "domain_row.writer = 'v2'",
		"event.parked_at IS NULL",
		"event.authority_group_generation <= group_row.generation",
		"event.cutover_generation <= domain_row.generation",
	} {
		if !strings.Contains(listOutboxAuthoritiesSQL, marker) {
			t.Fatalf("authority list SQL lacks %q", marker)
		}
	}
	for _, marker := range []string{
		"event.lease_until <= recovery_clock.observed_at",
		"FOR UPDATE OF event SKIP LOCKED", "OUTBOX_LEASE_EXPIRED",
		"event.lease_generation::text", "LIMIT $7",
	} {
		if !strings.Contains(recoverExpiredOutboxSQL, marker) {
			t.Fatalf("recovery SQL lacks %q", marker)
		}
	}
}

func TestPublisherSQLDocumentsParkOnlyEligibleStaleAuthority(t *testing.T) {
	t.Parallel()
	for _, marker := range []string{
		"event.parked_at IS NULL",
		"event.authority_group_generation < $4",
		"event.cutover_generation < $6",
		"event.delivery_state = 'pending'",
		"event.lease_until <= pg_catalog.clock_timestamp()",
		"platform.platform_async_park_stale_outbox_v1(",
		"candidate.event_id, pg_catalog.uuidv7()",
		"LIMIT $7",
	} {
		if !strings.Contains(parkStaleOutboxSQL, marker) {
			t.Fatalf("stale park SQL lacks %q", marker)
		}
	}
	for name, document := range map[string]string{
		"claim": claimOutboxSQL, "recovery": recoverExpiredOutboxSQL,
	} {
		if !strings.Contains(document, "event.parked_at IS NULL") {
			t.Fatalf("%s SQL can select parked outbox rows", name)
		}
	}
}
