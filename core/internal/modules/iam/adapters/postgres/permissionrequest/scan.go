package postgres

import (
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	"github.com/google/uuid"
)

type permissionRequestDocument struct {
	ID                  uuid.UUID  `json:"id"`
	Ordinal             int16      `json:"ordinal"`
	TargetKind          string     `json:"target_kind"`
	TargetID            uuid.UUID  `json:"target_id"`
	PermissionID        uuid.UUID  `json:"permission_id"`
	PermissionProtected bool       `json:"permission_protected"`
	PlannedBindingID    uuid.UUID  `json:"planned_binding_id"`
	GrantedBindingID    *uuid.UUID `json:"granted_binding_id"`
	State               string     `json:"state"`
	Revision            int64      `json:"revision"`
}

func scanPermissionRequestGroup(row permissionRequestRow) (permissionrequest.Group, error) {
	var group permissionrequest.Group
	var mode, state string
	var reviewerID, taskID *uuid.UUID
	var terminalAt *time.Time
	var requestsJSON []byte
	err := row.Scan(
		&group.ID, &group.RequesterID, &mode, &state, &reviewerID, &taskID,
		&group.Revision, &group.ExpiresAt, &terminalAt,
		&group.CreatedAt, &group.UpdatedAt, &requestsJSON,
	)
	if err != nil {
		return permissionrequest.Group{}, mapPermissionRequestError(err)
	}
	group.Mode = permissionrequest.Mode(mode)
	group.State = permissionrequest.State(state)
	if reviewerID != nil {
		group.ReviewerID = *reviewerID
	}
	if taskID != nil {
		group.ReviewTaskID = *taskID
	}
	if terminalAt != nil {
		group.TerminalAt = *terminalAt
	}
	var documents []permissionRequestDocument
	if err := json.Unmarshal(requestsJSON, &documents); err != nil {
		return permissionrequest.Group{}, permissionrequest.ErrUnavailable
	}
	group.Requests = make([]permissionrequest.Request, 0, len(documents))
	for _, document := range documents {
		request := permissionrequest.Request{
			ID: document.ID, Ordinal: document.Ordinal,
			Target: permissionrequest.Target{
				Kind: permissionrequest.TargetKind(document.TargetKind), ID: document.TargetID,
			},
			PermissionID:        document.PermissionID,
			PermissionProtected: document.PermissionProtected,
			PlannedBindingID:    document.PlannedBindingID,
			State:               permissionrequest.State(document.State), Revision: document.Revision,
		}
		if document.GrantedBindingID != nil {
			request.GrantedBindingID = *document.GrantedBindingID
		}
		group.Requests = append(group.Requests, request)
	}
	if group.Validate() != nil {
		return permissionrequest.Group{}, permissionrequest.ErrUnavailable
	}
	return group, nil
}
