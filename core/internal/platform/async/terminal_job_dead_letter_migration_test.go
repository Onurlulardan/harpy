package async

import (
	"strings"
	"testing"
)

func assertTerminalPayloadJCSOrder(t *testing.T, source string) {
	t.Helper()
	keys := []string{
		`'{"attempt":'`, `,"contract_version":`, `,"job_id":`, `,"kind":`,
		`,"payload_reference":`, `,"payload_revision":`, `,"payload_sha256":`,
		`,"reason_code":`,
	}
	position := -1
	for _, key := range keys {
		next := strings.Index(source[position+1:], key)
		if next < 0 {
			t.Fatalf("terminal DLQ JCS payload lacks ordered key %q", key)
		}
		position += next + 1
	}
}
