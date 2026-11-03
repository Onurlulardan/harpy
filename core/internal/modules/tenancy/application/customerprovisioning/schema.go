package customerprovisioning

import (
	"context"

	"github.com/google/uuid"
)

type SchemaProvisioner interface {
	Begin(context.Context, uuid.UUID, string, string, string) error
	Activate(context.Context, uuid.UUID) error
}
