package dashboardexport

import (
	"context"
	"errors"
	"time"
)

func (runner *RecoveryRunner) recoverMaintenanceUnknown(
	scan, force context.Context,
) (bool, error) {
	candidate, found, err := runner.store.NextMaintenanceUnknown(scan)
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
	claim, err := runner.store.ClaimMaintenanceUnknown(
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
	case UnknownBusy, UnknownResolved:
		return true, nil
	case UnknownClaimed:
		return true, runner.executeMaintenanceCleanup(force, claim)
	default:
		return true, ErrUnavailable
	}
}

func (runner *RecoveryRunner) executeMaintenanceCleanup(
	force context.Context, claim MaintenanceUnknownClaim,
) error {
	if claim.Mode == MaintenanceCleanupTerminal ||
		claim.Mode == MaintenanceCleanupCompleted {
		kind := UnknownResolutionAbsent
		if claim.Mode == MaintenanceCleanupCompleted &&
			claim.StoredEvidence.DeleteResponseSHA256 != [32]byte{} {
			kind = UnknownResolutionDeleted
		}
		return runner.resolveMaintenance(force, claim, maintenanceCleanupResult{
			kind: kind, evidence: claim.StoredEvidence})
	}
	if claim.Mode != MaintenanceCleanupArtifact {
		return ErrUnavailable
	}
	process, cancel := context.WithCancel(force)
	defer cancel()
	done := make(chan maintenanceCleanupResult, 1)
	go func() { done <- runner.cleanupMaintenanceArtifact(process, claim) }()
	heartbeats := time.NewTicker(recoveryHeartbeatInterval)
	defer heartbeats.Stop()
	for {
		select {
		case result := <-done:
			if errors.Is(result.err, ErrOutcomeUnknown) {
				_, err := runner.store.DeferMaintenanceUnknown(
					force, claim, runner.clock.Now())
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
			result.evidence.LocalStateSHA256 =
				claim.StoredEvidence.LocalStateSHA256
			result.evidence.LocalStateAt = claim.StoredEvidence.LocalStateAt
			return runner.resolveMaintenance(force, claim, result)
		case observedAt := <-heartbeats.C:
			lease, err := runner.store.CheckpointMaintenanceUnknown(
				force, claim, observedAt)
			if errors.Is(err, ErrDispatchLeaseStale) {
				cancel()
				<-done
				return nil
			}
			if err != nil || !lease.LeaseUntil.After(observedAt) {
				cancel()
				<-done
				if err != nil {
					return err
				}
				return ErrUnavailable
			}
		case <-force.Done():
			cancel()
			<-done
			return nil
		}
	}
}

func (runner *RecoveryRunner) resolveMaintenance(ctx context.Context,
	claim MaintenanceUnknownClaim, result maintenanceCleanupResult,
) error {
	receiptID, err := runner.ids.NewUUIDv7()
	if err != nil || !validUUIDv7(receiptID) {
		return errors.Join(ErrUnavailable, err)
	}
	verifiedAt := maintenanceEvidenceTime(result.evidence)
	resolvedAt := runner.clock.Now().UTC()
	if verifiedAt.IsZero() || verifiedAt.After(resolvedAt.Add(5*time.Second)) {
		return ErrUnavailable
	}
	minimum := claim.Candidate.UnknownAt.UTC().Add(time.Microsecond)
	if resolvedAt.Before(minimum) {
		resolvedAt = minimum
	}
	if resolvedAt.Before(verifiedAt) {
		resolvedAt = verifiedAt.UTC()
	}
	resolution := MaintenanceResolution{ReceiptID: receiptID, Kind: result.kind,
		TerminalState: claim.TerminalState, Evidence: result.evidence,
		ResolvedAt: resolvedAt}
	resolution.SHA256 = maintenanceResolutionDigest(
		resolution.Kind, resolution.TerminalState, resolution.Evidence, resolvedAt)
	_, err = runner.store.ResolveMaintenanceUnknown(ctx, claim, resolution)
	if errors.Is(err, ErrDispatchLeaseStale) {
		return nil
	}
	return err
}
