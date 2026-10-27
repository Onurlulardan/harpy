package domain

import "testing"

func TestProjectionVisibilityFailsClosedOnPermissionRevision(t *testing.T) {
	id := mustID(t, "018f9f3e-7c16-7a21-8b42-116a9d3c9870")
	item := ProjectionItem{ID: id, SourceID: id, SourceVersionID: id,
		Owner: SourceWorkflow, SourceRevision: 2, PermissionRevision: 7,
		EventSequence: 4, Summary: "approval"}
	if decision := DecideProjectionVisibility(item, true, 8); decision.Visible || !decision.RepairNeeded {
		t.Fatalf("stale permission decision = %+v", decision)
	}
	if decision := DecideProjectionVisibility(item, true, 7); !decision.Visible || decision.RepairNeeded {
		t.Fatalf("current permission decision = %+v", decision)
	}
}

func TestProjectionCursorParksGapsAndDeduplicates(t *testing.T) {
	cursor := ProjectionCursor{ThroughSequence: 2}
	if _, disposition, _ := cursor.Accept(2); disposition != ProjectionDuplicate {
		t.Fatalf("duplicate disposition = %q", disposition)
	}
	if _, disposition, _ := cursor.Accept(4); disposition != ProjectionGap {
		t.Fatalf("gap disposition = %q", disposition)
	}
	next, disposition, err := cursor.Accept(3)
	if err != nil || disposition != ProjectionApply || next.ThroughSequence != 3 {
		t.Fatalf("apply = %+v, %q, %v", next, disposition, err)
	}
}
