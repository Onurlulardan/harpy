package systemtransfer

import (
	"strings"
	"testing"
)

func TestDownloadExportArtifactPinsOwnedRoutine(t *testing.T) {
	if !strings.Contains(readExportArtifactSQL, "iam_system_record_export_content_v1") {
		t.Fatal("artifact SQL does not pin the owned routine")
	}
}
