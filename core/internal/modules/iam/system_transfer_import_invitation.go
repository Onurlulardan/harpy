package iam

import (
	"fmt"
	"time"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func newSystemTransferImportInvitation(
	settings config.IAMWorker,
	catalogs platformasync.WriteCatalogs,
	sealer InvitationDeliverySealer,
	roles systemTransferInvitationRoleResolver,
) (systemtransfer.InvitationCommand, error) {
	if settings.Validate() != nil || nilIdentityDependency(sealer) ||
		nilIdentityDependency(roles) {
		return nil, systemtransfer.ErrDependencies
	}
	keys, err := settings.OneTimeTokenKeyring()
	if err != nil {
		return nil, fmt.Errorf("load IAM import invitation keys: %w", err)
	}
	runtime, err := NewInvitationRuntime(catalogs, time.Now)
	if err != nil {
		return nil, err
	}
	application := &invitationHTTPApplication{
		runtime: runtime, keys: keys, sealer: sealer,
	}
	return newSystemTransferInvitationCommand(application, roles)
}
