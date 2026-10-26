package postgres

import (
	"context"
	"errors"
	"fmt"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (repository *workerRepository) ClaimJobs(
	ctx context.Context,
	authority JobAuthority,
	workerID uuid.UUID,
	limit int,
	catalog jobsv1.Catalog,
) (claims []JobClaim, resultErr error) {
	if ctx == nil {
		return nil, ErrWorkerContextInvalid
	}
	if !repository.valid() {
		return nil, ErrWorkerStoreUnavailable
	}
	if !authority.valid(repository.identity) {
		return nil, ErrJobAuthorityInvalid
	}
	if workerID == uuid.Nil {
		return nil, ErrJobClaimInvalid
	}
	if limit < 1 || limit > jobsv1.MaximumClaimBatch {
		return nil, ErrJobClaimLimit
	}
	policies, err := validatedClaimPolicies(catalog)
	if err != nil {
		return nil, err
	}
	transaction, err := repository.database.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, fmt.Errorf("begin job claim transaction: %w", err)
	}
	committed := false
	defer func() {
		if !committed {
			resultErr = errors.Join(resultErr, rollbackJobClaim(transaction, repository.cleanup))
		}
	}()
	if _, err := transaction.Exec(ctx, assertJobAuthoritySQL,
		authorityAssertionArguments(authority)...); err != nil {
		return nil, classifyJobAuthorityError(err)
	}
	rows, err := transaction.Query(
		ctx,
		claimJobsSQL,
		append(authorityArguments(authority), workerID, limit,
			policies.kinds, policies.revisions, policies.domains,
			policies.effects, policies.maximumAttempts, string(authority.mode))...,
	)
	if err != nil {
		return nil, fmt.Errorf("claim due jobs: %w", err)
	}
	claimed, err := scanClaimedJobRows(rows, limit)
	if err != nil {
		return nil, err
	}
	claimed, err = insertJobAttempts(ctx, transaction, claimed)
	if err != nil {
		return nil, err
	}
	claims = make([]JobClaim, len(claimed))
	for index := range claimed {
		claims[index], err = newJobClaim(claimed[index], authority, workerID)
		if err != nil {
			return nil, err
		}
	}
	if err := transaction.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit job claim transaction: %w", err)
	}
	committed = true
	return claims, nil
}

type claimPolicyArrays struct {
	kinds           []string
	revisions       []int32
	domains         []string
	effects         []string
	maximumAttempts []int32
}

func validatedClaimPolicies(catalog jobsv1.Catalog) (claimPolicyArrays, error) {
	validated, err := jobsv1.NewCatalog(catalog.Policies())
	if err != nil {
		return claimPolicyArrays{}, fmt.Errorf("validate job claim catalog: %w", err)
	}
	policies := validated.Policies()
	values := claimPolicyArrays{
		kinds: make([]string, 0, len(policies)), revisions: make([]int32, 0, len(policies)),
		domains: make([]string, 0, len(policies)), effects: make([]string, 0, len(policies)),
		maximumAttempts: make([]int32, 0, len(policies)),
	}
	for _, policy := range policies {
		values.kinds = append(values.kinds, policy.Kind)
		values.revisions = append(values.revisions, int32(policy.PayloadRevision))
		values.domains = append(values.domains, policy.OwnerDomain)
		values.effects = append(values.effects, string(policy.EffectClass))
		values.maximumAttempts = append(values.maximumAttempts, int32(policy.MaximumAttempts))
	}
	return values, nil
}

func authorityArguments(authority JobAuthority) []any {
	return []any{
		authority.scopeKind,
		authority.scopeID,
		authority.groupID,
		authority.groupGeneration,
		authority.domain,
		authority.generation,
	}
}

func authorityAssertionArguments(authority JobAuthority) []any {
	return append(authorityArguments(authority), string(authority.mode))
}

func classifyJobAuthorityError(err error) error {
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "P5601", "P5602", "P5603":
			return fmt.Errorf("%w: database authority assertion failed", ErrJobAuthorityStale)
		}
	}
	return fmt.Errorf("assert current job authority: %w", err)
}
