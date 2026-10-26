package link

import (
	"strings"
	"testing"
)

func TestFormRecordAdoptionReadyCheckpointIsDone(t *testing.T) {
	for name, statement := range map[string]string{
		"load":    loadFormRecordAdoptionSQL,
		"advance": advanceFormRecordAdoptionSQL,
	} {
		if !strings.Contains(statement,
			"lifecycle_state IN ('READY','SUCCEEDED')") {
			t.Fatalf("%s does not export READY as a completed inventory", name)
		}
	}
}
