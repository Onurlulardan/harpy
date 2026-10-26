package outboxdelivery

import (
	"context"
	"errors"
	"fmt"

	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
)

func (runner *Runner) runSession(ctx context.Context, broker Broker) error {
	for {
		worked, err := runner.scanAuthorities(ctx, broker)
		if err != nil {
			if cleanStop(err, ctx) {
				return nil
			}
			return err
		}
		if !worked {
			if err := waitContext(ctx, idlePollInterval); err != nil {
				return nil
			}
		}
	}
}

func (runner *Runner) scanAuthorities(ctx context.Context, broker Broker) (bool, error) {
	if _, admitted := runner.admission.Admit(); !admitted {
		return false, nil
	}
	token, err := runner.store.BeginOutboxAuthorityPage(authorityPageSize)
	if err != nil {
		return false, fmt.Errorf("begin outbox authority scan: %w", err)
	}
	worked := false
	for {
		if _, admitted := runner.admission.Admit(); !admitted {
			return worked, nil
		}
		page, err := runner.store.ListOutboxAuthorities(ctx, token)
		if err != nil {
			return worked, fmt.Errorf("list outbox authorities: %w", err)
		}
		for _, authority := range page.Authorities() {
			authorityWorked, err := runner.processAuthority(ctx, broker, authority)
			worked = worked || authorityWorked
			if err != nil && !errors.Is(err, asyncpostgres.ErrOutboxAuthorityStale) {
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
	broker Broker,
	authority asyncpostgres.OutboxAuthority,
) (bool, error) {
	if _, admitted := runner.admission.Admit(); !admitted {
		return false, nil
	}
	parked, err := runner.store.ParkStaleOutbox(
		ctx, authority, recoveryBatchSize,
	)
	if err != nil {
		return false, fmt.Errorf("park stale outbox events: %w", err)
	}
	if _, err := runner.store.RecoverExpiredOutbox(ctx, authority, recoveryBatchSize); err != nil {
		return false, fmt.Errorf("recover expired outbox leases: %w", err)
	}
	if _, admitted := runner.admission.Admit(); !admitted {
		return parked > 0, nil
	}
	claims, err := runner.store.ClaimOutbox(
		ctx, authority, runner.publisherID, claimBatchSize, runner.catalog,
	)
	if err != nil {
		return false, fmt.Errorf("claim outbox event: %w", err)
	}
	for _, claim := range claims {
		if err := runner.deliver(ctx, broker, claim); err != nil {
			return true, err
		}
	}
	return parked > 0 || len(claims) > 0, nil
}
