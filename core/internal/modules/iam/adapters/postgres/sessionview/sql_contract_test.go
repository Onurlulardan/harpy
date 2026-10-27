package sessionview

import (
	"strings"
	"testing"
)

func TestReadBindingUsesStableProfileSessionSuccessor(t *testing.T) {
	if !strings.Contains(readSQL, "iam_current_session_read_v6(") ||
		strings.Contains(readSQL, "iam_current_session_read_v5(") {
		t.Fatal("session read binding does not use v6")
	}
}
