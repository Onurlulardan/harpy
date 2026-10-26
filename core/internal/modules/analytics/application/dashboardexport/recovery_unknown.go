package dashboardexport

import (
	"context"
	"errors"
	"time"
)

const recoveryHeartbeatInterval = 20 * time.Second

type unknownCleanupResult struct {
	kind     UnknownResolutionKind
	evidence UnknownResolutionEvidence
	err      error
}

func (runner *RecoveryRunner) recoverUnknown(
	scan, force context.Context,
) (bool, error) {
	candidate, found, err := runner.store.NextUnknown(scan)
	if err != nil || !found {
		return false, err
	}
	if scan.Err() != nil {
		return false, nil
	}
	cleanupOwner, err := runner.ids.NewUUIDv7()
	if err != nil || !validUUIDv7(cleanupOwner) {
		return true, errors.Join(ErrUnavailable, err)
	}
	claim, err := runner.store.ClaimUnknown(
		force, candidate, cleanupOwner, runner.clock.Now())
	if errors.Is(err, ErrDispatchCandidateChanged) ||
		errors.Is(err, ErrDispatchLeaseStale) {
		return true, nil
	}
	if err != nil {
		if force.Err() != nil {
			return true, nil
		}
		return true, err
	}
	switch claim.Disposition {
	case UnknownBusy:
		return true, nil
	case UnknownResolved:
		return true, runner.reapResolved(force)
	case UnknownClaimed:
		return true, runner.executeUnknownCleanup(force, claim)
	default:
		return true, ErrUnavailable
	}
}

func (runner *RecoveryRunner) executeUnknownCleanup(
	force context.Context, claim UnknownCleanupClaim,
) error {
	process, cancel := context.WithCancel(force)
	defer cancel()
	done := make(chan unknownCleanupResult, 1)
	go func() { done <- runner.cleanupUnknownArtifact(process, claim) }()
	heartbeats := time.NewTicker(recoveryHeartbeatInterval)
	defer heartbeats.Stop()
	for {
		select {
		case result := <-done:
			if errors.Is(result.err, ErrOutcomeUnknown) {
				_, err := runner.store.DeferUnknown(force, claim, runner.clock.Now())
				if errors.Is(err, ErrDispatchLeaseStale) {
					return nil
				}
				return err
			}
			if result.err != nil {
				if force.Err() != nil {
					return nil
				}
				return result.err
			}
			return runner.resolveUnknown(force, claim, result)
		case observedAt := <-heartbeats.C:
			lease, err := runner.store.CheckpointUnknown(force, claim, observedAt)
			if errors.Is(err, ErrDispatchLeaseStale) {
				cancel()
				<-done
				return nil
			}
			if err != nil {
				cancel()
				<-done
				return err
			}
			if !lease.LeaseUntil.After(observedAt) {
				cancel()
				<-done
				return ErrUnavailable
			}
		case <-force.Done():
			cancel()
			<-done
			return nil
		}
	}
}

func (runner *RecoveryRunner) resolveUnknown(ctx context.Context,
	claim UnknownCleanupClaim, result unknownCleanupResult,
) error {
	receiptID, err := runner.ids.NewUUIDv7()
	if err != nil || !validUUIDv7(receiptID) {
		return errors.Join(ErrUnavailable, err)
	}
	resolvedAt := runner.clock.Now().UTC()
	minimum := claim.Candidate.UnknownAt.UTC().Add(time.Microsecond)
	verifiedAt := result.evidence.ReconciledAt
	if !result.evidence.DeletedAt.IsZero() {
		verifiedAt = result.evidence.DeletedAt
	}
	if verifiedAt.IsZero() || verifiedAt.Before(claim.Candidate.UnknownAt) ||
		verifiedAt.After(resolvedAt.Add(5*time.Second)) {
		return ErrUnavailable
	}
	if resolvedAt.Before(minimum) {
		resolvedAt = minimum
	}
	if resolvedAt.Before(verifiedAt) {
		resolvedAt = verifiedAt.UTC()
	}
	digest := unknownResolutionDigest(result.kind, result.evidence, resolvedAt)
	resolution := UnknownResolution{ReceiptID: receiptID, Kind: result.kind,
		SHA256: digest, Evidence: result.evidence, ResolvedAt: resolvedAt}
	_, err = runner.store.ResolveUnknown(ctx, claim, resolution)
	if errors.Is(err, ErrDispatchLeaseStale) {
		return nil
	}
	if err != nil {
		return err
	}
	return runner.reapResolved(ctx)
}
