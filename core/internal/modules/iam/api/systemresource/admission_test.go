package systemresource

import "testing"

func TestAdmissionBindsExactIAMLayoutGeneration(t *testing.T) {
	tenantID, _ := ParseID("018f47f4-7b1b-7cc0-8000-000000000001")
	actorID, _ := ParseID("018f47f4-7b1b-7cc0-8000-000000000002")
	actor, _ := NewActorEvidence(tenantID, actorID, 1, 2)
	admission, err := NewAdmission(tenantID, actor, 7, 7, 7)
	if err != nil || admission.Validate() != nil {
		t.Fatalf("admission = %#v: %v", admission, err)
	}
	for _, mismatch := range [][3]int64{{7, 6, 7}, {7, 7, 6}, {0, 0, 0}} {
		if _, err := NewAdmission(
			tenantID, actor, mismatch[0], mismatch[1], mismatch[2],
		); err == nil {
			t.Fatalf("mismatched generations accepted: %v", mismatch)
		}
	}
}
