package iam

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
)

func invitationHTTPView(record invitation.Invitation) InvitationView {
	return InvitationView{
		ID: record.ID, Email: record.DisplayValue, Role: record.RoleName,
		InvitedBy: record.InvitedBy, Status: strings.ToUpper(string(record.State)),
		ExpiresAt: record.ExpiresAt, CreatedAt: record.CreatedAt,
		Revision: record.Revision,
	}
}
