package domain

import "testing"

func TestTaskLabelChecklistAreTypedRelationalValues(t *testing.T) {
	id := mustID(t, "018f9f3e-7c16-7a21-8b42-116a9d3c9870")
	if !(Task{ID: id, BoardID: id, ColumnID: id, CreatorID: id, Kind: TaskCustom,
		Status: StatusOpen, Priority: PriorityMedium, Title: "Task",
		Position: SparseGap, RowVersion: 1}).Valid() {
		t.Fatal("valid task was rejected")
	}
	if !(Label{ID: id, BoardID: id, Name: "Urgent", Color: "#ff0000", RowVersion: 1}).Valid() {
		t.Fatal("valid label was rejected")
	}
	if !(ChecklistItem{ID: id, TaskID: id, Text: "Review", Position: SparseGap, RowVersion: 1}).Valid() {
		t.Fatal("valid checklist item was rejected")
	}
}
