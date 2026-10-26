package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
)

type EnvironmentIdentityRecord struct {
	Identity         recoveryset.EnvironmentIdentity
	DatabaseSystemID string
}

type EnvironmentIdentityReader interface {
	CurrentIdentity(context.Context, string) (EnvironmentIdentityRecord, error)
}
