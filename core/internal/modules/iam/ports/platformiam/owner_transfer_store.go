package platformiam

import (
	"context"

	platformdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/platformiam"
)

type OwnerTransferStore[S any] interface {
	CreateOwnerTransfer(
		context.Context, S, platformdomain.OwnerTransferPrepare,
	) (platformdomain.OwnerTransferPrepared, error)
	RecordOwnerTransferAck(
		context.Context, S, platformdomain.OwnerTransferAcknowledge,
	) (platformdomain.OwnerTransferPrepared, error)
	CompleteOwnerTransfer(
		context.Context, S, platformdomain.OwnerTransferCommit,
	) (platformdomain.OwnerTransferCommitted, error)
	ReserveOwnerTransferCompensation(
		context.Context, S, platformdomain.OwnerTransferCompensation,
	) (platformdomain.OwnerTransferPrepared, error)
	ExecuteOwnerTransferCompensation(
		context.Context, S, platformdomain.OwnerTransferAcknowledge,
	) (platformdomain.OwnerTransferCompensated, error)
}
