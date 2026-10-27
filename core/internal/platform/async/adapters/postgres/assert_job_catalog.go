package postgres

import (
	"context"
	"errors"
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func (repository *workerRepository) AssertJobCatalog(
	ctx context.Context,
	catalog jobsv1.Catalog,
) error {
	if ctx == nil {
		return ErrWorkerContextInvalid
	}
	if !repository.valid() {
		return ErrWorkerStoreUnavailable
	}
	policies, err := validatedClaimPolicies(catalog)
	if err != nil {
		return err
	}
	rows, err := repository.database.Query(
		ctx, assertLiveJobCatalogSQL, policies.kinds, policies.revisions,
		policies.domains, policies.effects, policies.maximumAttempts,
	)
	if err != nil {
		return fmt.Errorf("inspect live job catalog: %w", err)
	}
	defer rows.Close()
	var compatible bool
	if !rows.Next() {
		return fmt.Errorf("inspect live job catalog: %w", errors.Join(rows.Err(), ErrJobCatalogIncompatible))
	}
	if err := rows.Scan(&compatible); err != nil {
		return fmt.Errorf("scan live job catalog: %w", err)
	}
	if rows.Next() || rows.Err() != nil {
		return fmt.Errorf("inspect live job catalog cardinality: %w", ErrJobCatalogIncompatible)
	}
	if !compatible {
		return ErrJobCatalogIncompatible
	}
	return nil
}
