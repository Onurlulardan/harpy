package serviceprincipal

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type EvidencePort interface {
	Evidence(
		context.Context,
		transaction.TenantScope,
		uuid.UUID,
		int64,
	) (Principal, GrantRevision, error)
}
