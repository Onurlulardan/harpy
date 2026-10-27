package authorization

import (
	"errors"
	"testing"
)

func TestSnapshotOperationAdmissionAcceptsExactOneToFourOwnerSets(t *testing.T) {
	owners := []string{"forms", "crm", "workflow", "documents"}
	for count := 1; count <= 4; count++ {
		t.Run(owners[count-1], func(t *testing.T) {
			generations := make(map[string]int64, count)
			for _, owner := range owners[:count] {
				generations[owner] = int64(count)
			}
			spec := snapshotRecordSpec(t, count)
			owned, err := exactSnapshotGenerations(spec, generations)
			if err != nil || len(owned) != count {
				t.Fatalf("exact generations=%v error=%v", owned, err)
			}
			if _, err := NewSnapshotOperationAdmission().Issue(
				spec, generations,
			); err != nil {
				t.Fatalf("owner count %d rejected: %v", count, err)
			}
		})
	}
}

func TestSnapshotOperationAdmissionRejectsGenerationSetMismatch(t *testing.T) {
	tests := []struct {
		name        string
		ownerCount  int
		generations map[string]int64
	}{
		{"zero set", 1, nil},
		{"zero generation", 1, map[string]int64{"forms": 0}},
		{"extra generation", 1, map[string]int64{"forms": 1, "iam": 1}},
		{"missing generation", 2, map[string]int64{"forms": 1}},
		{"same size wrong key", 2, map[string]int64{"forms": 1, "iam": 1}},
	}
	admission := NewSnapshotOperationAdmission()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := admission.Issue(
				snapshotRecordSpec(t, test.ownerCount), test.generations,
			)
			if !errors.Is(err, errSnapshotOperationAdmission) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}

func TestSnapshotOperationAdmissionAcceptsAggregateRootOwner(t *testing.T) {
	if snapshotReadOperationID != "restquery.execute" {
		t.Fatalf("snapshot operation ID=%q", snapshotReadOperationID)
	}
	fixture := newAuthorizationQueryFixture(t)
	spec := authorizationAggregateSpec(t, fixture)
	if _, err := NewSnapshotOperationAdmission().Issue(
		spec, map[string]int64{"forms": 7},
	); err != nil {
		t.Fatalf("aggregate root rejected: %v", err)
	}
}

func TestSnapshotOperationAdmissionOwnsCallerGenerationMap(t *testing.T) {
	spec := snapshotRecordSpec(t, 3)
	caller := map[string]int64{"forms": 3, "crm": 5, "workflow": 7}
	owned, err := exactSnapshotGenerations(spec, caller)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewSnapshotOperationAdmission().Issue(spec, caller); err != nil {
		t.Fatal(err)
	}
	caller["forms"], caller["iam"] = 99, 11
	delete(caller, "workflow")
	if owned["forms"] != 3 || owned["workflow"] != 7 {
		t.Fatalf("owned generations changed: %v", owned)
	}
}

func TestSnapshotOperationAdmissionRejectsZeroIssuer(t *testing.T) {
	_, err := (SnapshotOperationAdmission{}).Issue(
		snapshotRecordSpec(t, 1), map[string]int64{"forms": 1},
	)
	if !errors.Is(err, errSnapshotOperationAdmission) {
		t.Fatalf("error=%v", err)
	}
}

func TestSnapshotOperationAdmissionIssuesOnlyExactFormsOwner(t *testing.T) {
	admission := NewSnapshotOperationAdmission()
	if _, err := admission.IssueForms(map[string]int64{"forms": 7}); err != nil {
		t.Fatalf("exact forms owner rejected: %v", err)
	}
	for _, generations := range []map[string]int64{
		nil,
		{"forms": 0},
		{"forms": 7, "iam": 9},
		{"iam": 7},
	} {
		if _, err := admission.IssueForms(generations); !errors.Is(
			err, errSnapshotOperationAdmission,
		) {
			t.Fatalf("invalid forms generations %v error=%v", generations, err)
		}
	}
}
