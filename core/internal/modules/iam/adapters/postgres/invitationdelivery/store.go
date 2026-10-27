package invitationdelivery

import invitationdeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitationdelivery"

type store struct {
	database deliveryDatabase
}

var _ invitationdeliveryapp.Store = (*store)(nil)
