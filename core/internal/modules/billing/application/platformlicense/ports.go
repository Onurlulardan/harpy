package platformlicense

import (
	"context"

	"github.com/google/uuid"
)

type Store interface {
	ReadLicense(context.Context, uuid.UUID) (License, error)
	Create(context.Context, RefreshCommand) (RefreshAccepted, error)
	Get(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (RefreshOperation, error)
}
