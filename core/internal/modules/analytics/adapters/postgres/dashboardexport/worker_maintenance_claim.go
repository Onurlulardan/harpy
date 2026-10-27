package dashboardexport

import (
	"context"
	"time"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

func (storage *WorkerStore) ClaimMaintenanceUnknown(
	ctx context.Context, candidate exportapp.MaintenanceUnknownCandidate,
	cleanupOwner uuid.UUID, observedAt time.Time,
) (exportapp.MaintenanceUnknownClaim, error) {
	if !validMaintenanceCandidate(candidate) || cleanupOwner.Version() != 7 ||
		observedAt.IsZero() {
		return exportapp.MaintenanceUnknownClaim{}, exportapp.ErrInvalid
	}
	result := exportapp.MaintenanceUnknownClaim{}
	err := storage.inTenant(ctx, candidate.TenantID, func(boundContext context.Context,
		bound *workerBound,
	) error {
		schema, err := bound.currentSchema(boundContext)
		if err != nil || schema != candidate.TenantSchema {
			return exportapp.ErrUnavailable
		}
		stored := maintenanceClaimRow{}
		arguments := append(maintenanceCandidateArguments(schema, candidate),
			cleanupOwner, observedAt)
		err = bound.queryRow(boundContext, workerMaintenanceUnknownClaimSQL,
			arguments...).Scan(stored.targets()...)
		if err != nil {
			return err
		}
		result, err = stored.decode(candidate, cleanupOwner)
		return err
	})
	if err != nil {
		return exportapp.MaintenanceUnknownClaim{}, mapDispatchError(err)
	}
	return result, nil
}

func maintenanceCandidateArguments(schema string,
	candidate exportapp.MaintenanceUnknownCandidate,
) []any {
	return []any{schema, candidate.Kind, candidate.TenantID, candidate.TenantSchema,
		candidate.ExportID, candidate.AuthorityGroupID,
		candidate.AuthorityGroupGeneration, candidate.CutoverGeneration,
		candidate.JobID, candidate.AttemptID,
		candidate.AttemptNumber, candidate.LeaseGeneration,
		candidate.OriginalLeaseOwnerID, candidate.EffectID,
		candidate.EffectRequestSHA256[:], candidate.UnknownAt}
}

func validMaintenanceCandidate(candidate exportapp.MaintenanceUnknownCandidate) bool {
	maximum := int64(3)
	if candidate.Kind == "analytics.dashboard.export.expire" {
		maximum = 8
	} else if candidate.Kind != "analytics.dashboard.export.delete" {
		return false
	}
	return candidate.TenantID != uuid.Nil &&
		workerTenantSchemaPattern.MatchString(candidate.TenantSchema) &&
		workerTenantSchemaPattern.MatchString(candidate.AuthorityGroupID) &&
		candidate.ExportID.Version() == 7 && candidate.JobID.Version() == 7 &&
		candidate.AttemptID.Version() == 7 &&
		candidate.OriginalLeaseOwnerID.Version() == 7 &&
		candidate.EffectID.Version() == 5 &&
		candidate.AuthorityGroupGeneration > 0 && candidate.CutoverGeneration > 0 &&
		candidate.AttemptNumber >= 1 && candidate.AttemptNumber <= maximum &&
		candidate.LeaseGeneration == candidate.AttemptNumber &&
		candidate.EffectRequestSHA256 != [32]byte{} && !candidate.UnknownAt.IsZero()
}

func validMaintenanceClaim(claim exportapp.MaintenanceUnknownClaim) bool {
	return claim.Disposition == exportapp.UnknownClaimed &&
		validMaintenanceCandidate(claim.Candidate) && claim.CleanupGeneration >= 1 &&
		claim.DeleteAttemptGeneration >= 0 && claim.DeleteAttemptGeneration <= 3 &&
		(claim.Mode != exportapp.MaintenanceCleanupArtifact ||
			claim.DeleteAttemptGeneration >= 1) &&
		claim.CleanupLeaseOwnerID.Version() == 7 &&
		claim.CleanupLeaseUntil.After(claim.Candidate.UnknownAt) &&
		claim.StoredEvidence.LocalStateSHA256 != [32]byte{} &&
		!claim.StoredEvidence.LocalStateAt.IsZero() &&
		validMaintenanceTerminal(claim.Candidate.Kind, claim.TerminalState)
}
