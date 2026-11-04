package transaction

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

func lockOperationAuthority(
	ctx context.Context,
	database *postgres.RuntimeTransaction,
	tenantID uuid.UUID,
	operation VerifiedOperation,
) error {
	if !operation.valid() {
		return ErrWriterAuthority
	}
	for _, domain := range operation.domains {
		if err := verifyOperationDomain(
			ctx, database, tenantID, operation.mode, domain,
		); err != nil {
			return mapPreparationError(err)
		}
	}
	return nil
}

func verifyOperationDomain(
	ctx context.Context,
	database *postgres.RuntimeTransaction,
	tenantID uuid.UUID,
	mode ScopeMode,
	domain domainExpectation,
) error {
	if mode == SnapshotQuery {
		return database.VerifyDomainCutover(
			ctx, tenantID, domain.name, domain.generation,
		)
	}
	return nil
}
