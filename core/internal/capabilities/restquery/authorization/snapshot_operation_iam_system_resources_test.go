package authorization

import (
	"errors"
	"testing"
)

func TestIAMSystemResourcesSnapshotAdmissionAcceptsExactPairedGeneration(
	t *testing.T,
) {
	_, err := NewSnapshotOperationAdmission().IssueIAMSystemResources(
		map[string]int64{"iam": 17, "system_resource_layout": 17},
	)
	if err != nil {
		t.Fatalf("paired IAM system-resource snapshot rejected: %v", err)
	}
}

func TestIAMSystemResourcesSnapshotAdmissionRejectsNonExactSets(t *testing.T) {
	invalid := []map[string]int64{
		nil,
		{"iam": 1},
		{"iam": 0, "system_resource_layout": 0},
		{"iam": 1, "system_resource_layout": 2},
		{"iam": 1, "system_resource_layout": 1, "forms": 1},
		{"iam": 1, "forms": 1},
	}
	admission := NewSnapshotOperationAdmission()
	for _, generations := range invalid {
		if _, err := admission.IssueIAMSystemResources(generations); !errors.Is(
			err, errSnapshotOperationAdmission,
		) {
			t.Fatalf("invalid generations %v error=%v", generations, err)
		}
	}
}

func TestIAMSystemResourcesSnapshotAdmissionRejectsZeroIssuer(t *testing.T) {
	_, err := (SnapshotOperationAdmission{}).IssueIAMSystemResources(
		map[string]int64{"iam": 3, "system_resource_layout": 3},
	)
	if !errors.Is(err, errSnapshotOperationAdmission) {
		t.Fatalf("error=%v", err)
	}
}
