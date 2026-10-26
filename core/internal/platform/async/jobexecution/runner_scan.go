package jobexecution

import (
	"context"
	"errors"
	"fmt"
	"time"

	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

type jobClaim struct{ value asyncpostgres.JobClaim }

func (runner *Runner) scanLoop(ctx context.Context, group *executionGroup) error {
	for {
		worked, err := runner.scanAuthorities(ctx, group)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if !worked {
			select {
			case <-ctx.Done():
				return nil
			case err := <-group.failures:
				return err
			case <-group.released:
			case <-time.After(idlePollInterval):
			}
		}
	}
}

func (runner *Runner) scanAuthorities(
	ctx context.Context,
	group *executionGroup,
) (bool, error) {
	if _, admitted := runner.admission.Admit(); !admitted {
		return false, nil
	}
	parked, err := runner.store.ReconcileStaleJobs(ctx, recoveryBatchSize)
	if err != nil {
		return false, fmt.Errorf("reconcile stale jobs: %w", err)
	}
	token, err := runner.store.BeginJobAuthorityPage(authorityPageSize)
	if err != nil {
		return parked > 0, fmt.Errorf("begin job authority scan: %w", err)
	}
	worked := parked > 0
	for {
		if _, admitted := runner.admission.Admit(); !admitted {
			return worked, nil
		}
		page, err := runner.store.ListJobAuthorities(ctx, token)
		if err != nil {
			return worked, fmt.Errorf("list job authorities: %w", err)
		}
		for _, authority := range page.Authorities() {
			authorityWorked, err := runner.processAuthority(ctx, group, authority)
			worked = worked || authorityWorked
			if err != nil {
				return worked, err
			}
		}
		next, present := page.Next()
		if !present {
			return worked, nil
		}
		token = next
	}
}

func (runner *Runner) processAuthority(
	ctx context.Context,
	group *executionGroup,
	authority asyncpostgres.JobAuthority,
) (bool, error) {
	if _, admitted := runner.admission.Admit(); !admitted {
		return false, nil
	}
	recovered, err := runner.store.RecoverExpiredJobs(ctx, authority, recoveryBatchSize)
	if err != nil {
		if errors.Is(err, asyncpostgres.ErrJobAuthorityStale) {
			return false, nil
		}
		return false, fmt.Errorf("recover expired job leases: %w", err)
	}
	available := maximumConcurrent - len(group.slots)
	if available == 0 {
		return recovered.Total() > 0, nil
	}
	if available > claimBatchSize {
		available = claimBatchSize
	}
	if _, admitted := runner.admission.Admit(); !admitted {
		return recovered.Total() > 0, nil
	}
	claims, err := runner.store.ClaimJobs(
		ctx, authority, runner.workerID, available, runner.registry.Catalog(),
	)
	if err != nil {
		if errors.Is(err, asyncpostgres.ErrJobAuthorityStale) {
			return recovered.Total() > 0, nil
		}
		return recovered.Total() > 0, fmt.Errorf("claim due jobs: %w", err)
	}
	for _, claim := range claims {
		group.launch(runner, jobClaim{value: claim})
	}
	return recovered.Total() > 0 || len(claims) > 0, nil
}
