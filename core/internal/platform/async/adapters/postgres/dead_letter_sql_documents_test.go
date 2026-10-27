package postgres

import (
	"strings"
	"testing"
)

func TestDeadLetterSQLDocumentsArePresentAndNeverResetOrDeleteState(t *testing.T) {
	documents := map[string]string{
		"resolve_authority": resolveDeadLetterAuthoritySQL,
		"insert_event":      insertDeadLetterEventSQL,
		"list_roots":        listDeadLetterRootsSQL,
		"lock_root":         lockDeadLetterRootSQL,
		"insert_action":     insertDeadLetterActionSQL,
		"event_retry":       insertDeadLetterEventRetrySQL,
		"job_retry":         insertDeadLetterJobRetrySQL,
	}
	for name, document := range documents {
		normalized := strings.ToUpper(strings.TrimSpace(document))
		if normalized == "" {
			t.Fatalf("%s SQL document is empty", name)
		}
		if strings.Contains(normalized, "DELETE ") ||
			strings.Contains(normalized, "UPDATE PLATFORM.JOBS") ||
			strings.Contains(normalized, "UPDATE PLATFORM.DEAD_LETTERS") {
			t.Fatalf("%s SQL can reset or delete authoritative state", name)
		}
	}
}
