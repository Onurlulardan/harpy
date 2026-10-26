package transaction

import (
	"context"
	"slices"

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
	if operation.mode == Command {
		for _, group := range operation.groups {
			if err := lockOperationGroup(ctx, database, tenantID, group); err != nil {
				return err
			}
		}
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

func lockOperationGroup(
	ctx context.Context,
	database *postgres.RuntimeTransaction,
	tenantID uuid.UUID,
	group groupExpectation,
) error {
	if err := database.LockCutoverGroup(
		ctx, tenantID, group.id, group.generation,
	); err != nil {
		return mapPreparationError(err)
	}
	members, err := database.ListCutoverGroupMembers(ctx, tenantID, group.id)
	expected, exists := tenantAuthorityGroupMembers(group.id)
	if err != nil || !exists || !slices.Equal(members, expected) {
		return ErrWriterAuthority
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
	return database.LockDomainCutover(
		ctx, tenantID, domain.name, domain.generation,
	)
}
