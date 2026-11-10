package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
		detail := incompatibleJobDetail(ctx, repository, policies)
		if detail == "" {
			return ErrJobCatalogIncompatible
		}
		return fmt.Errorf("%w: %s", ErrJobCatalogIncompatible, detail)
	}
	return nil
}

func incompatibleJobDetail(
	ctx context.Context,
	repository *workerRepository,
	policies claimPolicyArrays,
) string {
	rows, err := repository.database.Query(
		ctx, listIncompatibleJobCatalogSQL, policies.kinds, policies.revisions,
		policies.domains, policies.effects, policies.maximumAttempts,
	)
	if err != nil {
		return ""
	}
	defer rows.Close()
	var mismatches []string
	for rows.Next() {
		var kind, domain, effect string
		var revision, attempts int64
		if rows.Scan(&kind, &revision, &domain, &effect, &attempts) != nil {
			return ""
		}
		mismatches = append(mismatches, fmt.Sprintf(
			"%s revision %d domain %s effect %s attempts %d",
			kind, revision, domain, effect, attempts,
		))
	}
	if rows.Err() != nil {
		return ""
	}
	return strings.Join(mismatches, "; ")
}
