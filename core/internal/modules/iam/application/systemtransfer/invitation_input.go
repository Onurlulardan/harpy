package systemtransfer

import (
	"encoding/json"
	"strings"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func invitationInput(
	values publicresource.FieldValues,
	idempotencyKey string,
) (InvitationInput, error) {
	members, err := values.Values()
	if err != nil {
		return InvitationInput{}, ErrInvitationInvalid
	}
	emailID, _ := publicresource.ParseID(publicresource.UserEmailFieldID)
	roleIDsID, _ := publicresource.ParseID(publicresource.UserRoleIDsFieldID)
	var email string
	if json.Unmarshal(members[emailID], &email) != nil || strings.TrimSpace(email) != email ||
		email == "" {
		return InvitationInput{}, ErrInvitationInvalid
	}
	var roles []string
	if json.Unmarshal(members[roleIDsID], &roles) != nil || len(roles) != 1 {
		return InvitationInput{}, ErrInvitationInvalid
	}
	roleID, parseErr := publicresource.ParseID(roles[0])
	input := InvitationInput{Email: email, RoleID: roleID, IdempotencyKey: idempotencyKey}
	if parseErr != nil || input.Validate() != nil {
		return InvitationInput{}, ErrInvitationInvalid
	}
	return input, nil
}
