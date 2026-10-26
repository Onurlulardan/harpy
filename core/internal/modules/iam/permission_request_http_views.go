package iam

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	"github.com/google/uuid"
)

func permissionRequestGroupView(
	tenantID uuid.UUID,
	group permissionrequest.Group,
) PermissionRequestGroupView {
	view := PermissionRequestGroupView{
		ID: group.ID.String(), RequesterID: group.RequesterID.String(),
		Mode: string(group.Mode), State: string(group.State), Revision: group.Revision,
		ExpiresAt:           group.ExpiresAt.UTC().Format(time.RFC3339Nano),
		CreatedAt:           group.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:           group.UpdatedAt.UTC().Format(time.RFC3339Nano),
		StepUpPurposeSHA256: permissionRequestPurposeWire(tenantID, group.ID),
		Items:               make([]PermissionRequestItemView, len(group.Requests)),
	}
	if group.ReviewerID != uuid.Nil {
		value := group.ReviewerID.String()
		view.ReviewerID = &value
	}
	if group.ReviewTaskID != uuid.Nil {
		value := group.ReviewTaskID.String()
		view.ReviewTaskID = &value
	}
	if !group.TerminalAt.IsZero() {
		view.TerminalAt = group.TerminalAt.UTC().Format(time.RFC3339Nano)
	}
	for index, item := range group.Requests {
		itemView := PermissionRequestItemView{
			ID: item.ID.String(), Ordinal: item.Ordinal,
			TargetKind: string(item.Target.Kind), TargetID: item.Target.ID.String(),
			PermissionID:        item.PermissionID.String(),
			PermissionProtected: item.PermissionProtected,
			PlannedBindingID:    item.PlannedBindingID.String(),
			State:               string(item.State), Revision: item.Revision,
		}
		if item.GrantedBindingID != uuid.Nil {
			value := item.GrantedBindingID.String()
			itemView.GrantedBindingID = &value
		}
		view.Items[index] = itemView
	}
	return view
}

func permissionRequestRefView(
	tenantID uuid.UUID,
	ref permissionrequest.GroupRef,
) PermissionRequestRefView {
	return PermissionRequestRefView{
		ID: ref.ID.String(), State: string(ref.State), Revision: ref.Revision,
		StepUpPurposeSHA256: permissionRequestPurposeWire(tenantID, ref.ID),
	}
}
