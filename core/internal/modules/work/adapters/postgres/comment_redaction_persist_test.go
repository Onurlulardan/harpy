package postgres

import (
	"strings"
	"testing"
)

func TestCommentRedactionPersistSQLUsesVisibleCAS(t *testing.T) {
	if !strings.Contains(redactCommentSQL, "state = 'VISIBLE'") {
		t.Fatal("missing redaction CAS")
	}
}
