package domain

import "errors"

var ErrProjectionEvent = errors.New("work projection event is invalid")

type SourceOwner string

const (
	SourceWork     SourceOwner = "WORK"
	SourceWorkflow SourceOwner = "WORKFLOW"
	SourceForms    SourceOwner = "FORMS"
)

func (owner SourceOwner) Valid() bool {
	return owner == SourceWork || owner == SourceWorkflow || owner == SourceForms
}

type ProjectionItem struct {
	ID, SourceID, SourceVersionID                     ID
	Owner                                             SourceOwner
	SourceRevision, PermissionRevision, EventSequence int64
	Summary                                           string
}

func (item ProjectionItem) Valid() bool {
	return item.ID.Valid() && item.SourceID.Valid() && item.SourceVersionID.Valid() &&
		item.Owner.Valid() && item.SourceRevision > 0 && item.PermissionRevision > 0 &&
		item.EventSequence > 0 && len(item.Summary) <= 500
}

type VisibilityDecision struct {
	Visible      bool
	RepairNeeded bool
}

func DecideProjectionVisibility(item ProjectionItem, allowed bool, currentPermissionRevision int64) VisibilityDecision {
	if !item.Valid() || currentPermissionRevision < 1 ||
		item.PermissionRevision != currentPermissionRevision {
		return VisibilityDecision{RepairNeeded: true}
	}
	return VisibilityDecision{Visible: allowed, RepairNeeded: !allowed}
}

type ProjectionCursor struct {
	ThroughSequence int64
}

type ProjectionDisposition string

const (
	ProjectionApply     ProjectionDisposition = "APPLY"
	ProjectionDuplicate ProjectionDisposition = "DUPLICATE"
	ProjectionGap       ProjectionDisposition = "GAP"
)

func (cursor ProjectionCursor) Accept(sequence int64) (ProjectionCursor, ProjectionDisposition, error) {
	if cursor.ThroughSequence < 0 || sequence < 1 {
		return cursor, "", ErrProjectionEvent
	}
	if sequence <= cursor.ThroughSequence {
		return cursor, ProjectionDuplicate, nil
	}
	if sequence != cursor.ThroughSequence+1 {
		return cursor, ProjectionGap, nil
	}
	return ProjectionCursor{ThroughSequence: sequence}, ProjectionApply, nil
}
