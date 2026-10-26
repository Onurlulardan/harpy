package domain

import "testing"

func TestResourceLinkRequiresStableOwnerIdentity(t *testing.T) {
	id := mustID(t, "018f9f3e-7c16-7a21-8b42-116a9d3c9870")
	link := ResourceLink{ID: id, WorkResourceID: id, TargetID: id,
		TargetVersionID: id, TargetOwner: SourceForms, TargetKind: "FORM_RECORD",
		ReferenceEpoch: 1, PermissionRevision: 1, RowVersion: 1}
	if !link.Valid() {
		t.Fatal("stable resource link was rejected")
	}
	link.TargetVersionID = ID{}
	if link.Valid() {
		t.Fatal("unversioned resource link was accepted")
	}
}
