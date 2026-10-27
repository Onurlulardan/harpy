package systemtransfer

import (
	"strings"
	"testing"
)

func TestBeginArtifactEffectPinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(outboundEffectStartSQL,
		"iam_system_record_export_worker_effect_v1") {
		t.Fatal("export effect start does not pin the owned routine")
	}
}
