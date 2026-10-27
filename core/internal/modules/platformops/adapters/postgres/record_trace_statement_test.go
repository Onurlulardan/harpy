package postgres

import (
	"strings"
	"testing"
)

func TestRecordTraceStatementMatchesRoutineSignature(t *testing.T) {
	if !strings.Contains(recordTraceSQL, "$1::uuid,$2::uuid,$3::text") {
		t.Fatal("record trace statement does not match the routine UUID/text prefix")
	}
	if strings.Contains(recordTraceSQL, "$3::uuid") {
		t.Fatal("correlation ID must not be cast to UUID")
	}
}
