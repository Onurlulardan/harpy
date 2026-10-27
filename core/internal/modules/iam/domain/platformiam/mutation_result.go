package platformiam

type MutationResult struct {
	SubjectID, RoleID, SubjectRoleBindingID, RolePermissionBindingID ID
	TargetUserID                                                     ID
	TargetAuthorizationVersion                                       int64
	PlatformClock                                                    int64
	Replayed                                                         bool
}

func (result MutationResult) Validate(target ID) error {
	if !result.SubjectID.Valid() || !result.RoleID.Valid() || result.TargetUserID != target ||
		!result.SubjectRoleBindingID.Valid() || !result.RolePermissionBindingID.Valid() ||
		result.TargetAuthorizationVersion < 1 || result.PlatformClock < 1 || !target.Valid() {
		return ErrStorage
	}
	return nil
}
