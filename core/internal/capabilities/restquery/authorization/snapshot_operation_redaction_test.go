package authorization

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestSnapshotOperationAdmissionIsRedacted(t *testing.T) {
	admission := NewSnapshotOperationAdmission()
	if admission.Validate() != nil || fmt.Sprint(admission) != authorizationRedaction ||
		admission.LogValue().String() != authorizationRedaction {
		t.Fatal("snapshot admission is invalid or leaked")
	}
	if _, err := json.Marshal(admission); err == nil {
		t.Fatal("snapshot admission allowed direct JSON marshal")
	}
}
