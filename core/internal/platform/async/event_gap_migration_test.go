package async

import (
	"bytes"
	"strings"
	"testing"
)

func TestEventGapMigrationsArePairedAndFreezeRecoveryContract(t *testing.T) {
	t.Parallel()
	documents, err := MigrationDocuments()
	if err != nil {
		t.Fatal(err)
	}
	byOrigin := map[string]map[string][]byte{
		"fresh": {}, "v1_adopted": {},
	}
	expectedNames := []string{
		"20260715021800_expand_event_gap_recovery_v1",
		"20260715021810_create_event_gap_job_readiness_v1",
		"20260715021820_create_event_gap_requeue_guard_v1",
		"20260715021825_create_live_event_gap_guard_v1",
		"20260715021830_allow_event_gap_outbox_requeue_v1",
		"20260715021832_guard_live_event_gap_retry_v2",
		"20260715021833_guard_live_event_gap_redaction_v2",
		"20260715021834_create_live_gap_safe_redaction_v2",
		"20260715021835_create_event_gap_action_v1",
		"20260715021836_retire_unsafe_redaction_v1",
		"20260715021838_create_event_gap_completion_v1",
		"20260715021840_create_event_gap_reconciliation_v1",
		"20260715021850_grant_event_gap_recovery_v1",
		"20260715021852_verify_stale_job_disposition_replay_v1",
		"20260715021854_validate_stale_job_disposition_v1",
		"20260715021855_create_live_work_safe_dead_letter_purge_v2",
		"20260715021856_create_stale_event_gap_reissue_v1",
		"20260715021857_create_stale_job_disposition_evidence_v1",
		"20260715021858_create_stale_job_disposition_v1",
		"20260715021859_create_outbox_authority_current_v1",
		"20260715021900_unique_terminal_job_dead_letter_root_v1",
		"20260715021910_record_terminal_job_dead_letter_v1",
		"20260715021920_bind_terminal_job_dead_letter_v1",
		"20260715021930_create_event_gap_job_readiness_v2",
		"20260715021940_create_event_gap_reconciliation_v2",
		"20260715021950_grant_stale_disposition_and_gap_v2",
	}
	expected := make(map[string]struct{}, len(expectedNames))
	for _, name := range expectedNames {
		expected[name] = struct{}{}
	}
	for _, document := range documents {
		name := document.Path()[strings.LastIndex(document.Path(), "/")+1:]
		name = strings.TrimSuffix(name, ".fresh.sql")
		name = strings.TrimSuffix(name, ".v1_adopted.sql")
		_, expectedName := expected[name]
		if len(name) >= 14 && name[:14] >= "20260715021800" &&
			name[:14] <= "20260715021950" && !expectedName {
			t.Fatalf("unexpected event-gap migration in frozen range: %s", name)
		}
		if !expectedName {
			continue
		}
		for origin := range byOrigin {
			if strings.Contains(document.Path(), "/"+origin+"/") {
				byOrigin[origin][name] = document.Content()
			}
		}
	}
	if len(byOrigin["fresh"]) != len(expected) ||
		len(byOrigin["v1_adopted"]) != len(expected) {
		t.Fatalf("event gap migration counts = fresh %d adopted %d",
			len(byOrigin["fresh"]), len(byOrigin["v1_adopted"]))
	}
	for name, fresh := range byOrigin["fresh"] {
		adopted, ok := byOrigin["v1_adopted"][name]
		if !ok || !bytes.Equal(fresh, adopted) {
			t.Fatalf("event gap migration pair drifted: %s", name)
		}
	}
	disposition := string(byOrigin["fresh"]["20260715021858_create_stale_job_disposition_v1"])
	for _, marker := range []string{
		"PERFORM platform.platform_async_assert_current_authority_v1(",
		"WHERE job.id = p_job_id FOR UPDATE;",
	} {
		if !strings.Contains(disposition, marker) {
			t.Fatalf("stale disposition lost transaction revalidation %q", marker)
		}
	}
	combined := migrationOriginSource(byOrigin["fresh"])
	for _, marker := range []string{
		"'sequence_gap'", "platform_async_event_gap_job_ready_v1",
		"platform_async_is_gap_requeue_v1", "action.recorded_at = p_new.next_attempt_at",
		"platform_async_record_event_gap_action_v1", "platform_async_complete_event_gap_job_v1",
		"platform_async_guard_outbox_transition_v2", "CREATE OR REPLACE TRIGGER outbox_transition_guard",
		"FOR EACH ROW WHEN (NEW.parked_at IS NULL)",
		"platform_async_has_live_event_gap_v1", "NOT platform.platform_async_has_live_event_gap_v1(OLD)",
		"platform_async_validate_dead_letter_chain_v2", "live event gap evidence cannot be redacted",
		"platform_async_redact_outbox_payloads_v2", "platform_async_redact_dead_letter_payloads_v2",
		"unsafe outbox redaction v1 is retired; use v2",
		"unsafe dead-letter redaction v1 is retired; use v2",
		"platform_async_event_gap_job_ready_v2", "platform_async_reconcile_event_gap_v2",
		"platform_async_dispose_stale_job_v1", "stale-job-disposition.v1",
		"platform_async_validate_stale_job_disposition_v1",
		"platform_async_verify_stale_job_disposition_replay_v1", "FOR SHARE;",
		"platform_async_purge_dead_letters_v2", "TO harpy_runtime;",
		"FROM PUBLIC;", "ERRCODE = 'P5697'",
	} {
		if !strings.Contains(combined, marker) {
			t.Fatalf("event gap migrations lack %q", marker)
		}
	}
	if strings.Contains(combined, "CREATE OR REPLACE FUNCTION") ||
		strings.Contains(combined, "LANGUAGE sql") ||
		strings.Contains(combined, "TO harpy_v2_runtime") ||
		strings.Contains(combined, "GRANT EXECUTE ON FUNCTION platform.platform_async_is_gap_requeue") {
		t.Fatal("event gap helper privilege escaped the logical runtime role")
	}
}

func migrationOriginSource(documents map[string][]byte) string {
	var source strings.Builder
	for _, document := range documents {
		source.Write(document)
	}
	return source.String()
}
