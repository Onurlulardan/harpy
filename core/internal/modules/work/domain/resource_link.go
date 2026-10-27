package domain

type ResourceLink struct {
	ID, WorkResourceID, TargetID, TargetVersionID  ID
	TargetOwner                                    SourceOwner
	TargetKind                                     string
	ReferenceEpoch, PermissionRevision, RowVersion int64
}

func (link ResourceLink) Valid() bool {
	return link.ID.Valid() && link.WorkResourceID.Valid() && link.TargetID.Valid() &&
		link.TargetVersionID.Valid() && link.TargetOwner.Valid() &&
		link.TargetOwner != SourceWork && link.TargetKind != "" &&
		len(link.TargetKind) <= 64 && link.ReferenceEpoch > 0 &&
		link.PermissionRevision > 0 && link.RowVersion > 0
}
