package initialadminsetup

import (
	"context"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/lifecyclecredential"
	"github.com/google/uuid"
)

type StageStore interface {
	Stage(context.Context, Stage) error
}

type DeliverySealer interface {
	SealLifecycleCredential(context.Context, uuid.UUID, uuid.UUID, uuid.UUID,
		uuid.UUID, uuid.UUID, lifecycledeliveryv1.Purpose, []byte,
	) (lifecycledeliveryv1.Envelope, error)
}

type Dependencies struct {
	Store    StageStore
	Keys     lifecyclecredential.OneTimeKeyring
	Delivery DeliverySealer
}
