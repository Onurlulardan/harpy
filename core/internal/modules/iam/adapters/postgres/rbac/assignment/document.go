package assignmentpostgres

import (
	"encoding/json"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/google/uuid"
)

type memberDocument struct {
	Ordinal      int       `json:"ordinal"`
	ActionID     uuid.UUID `json:"action_id"`
	PermissionID uuid.UUID `json:"permission_id"`
	BindingID    uuid.UUID `json:"binding_id"`
}

func assignmentDocument(command rbacdomain.AssignmentMutation) ([]byte, error) {
	members := make([]memberDocument, len(command.Items))
	for index, item := range command.Items {
		members[index] = memberDocument{
			Ordinal: item.Ordinal, ActionID: databaseID(item.ActionID),
			PermissionID: databaseID(item.PermissionID),
			BindingID:    databaseID(item.BindingID),
		}
	}
	return json.Marshal(members)
}
