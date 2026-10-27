package dashboardexport

import (
	"context"

	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func (storage *WorkerStore) ResolveMaintenanceUnknown(
	ctx context.Context, claim exportapp.MaintenanceUnknownClaim,
	resolution exportapp.MaintenanceResolution,
) (exportapp.State, error) {
	if !validMaintenanceClaim(claim) || !validMaintenanceResolution(claim, resolution) {
		return "", exportapp.ErrInvalid
	}
	platformCandidate := platformMaintenanceCandidate(claim.Candidate)
	effectReceiptID, err := platformasync.AnalyticsDashboardExportMaintenanceReceiptID(
		platformCandidate)
	if err != nil {
		return "", exportapp.ErrInvalid
	}
	effectReceiptSHA, err :=
		platformasync.AnalyticsDashboardExportMaintenanceReceiptSHA256(
			platformCandidate, effectReceiptID, string(resolution.TerminalState),
			resolution.ResolvedAt)
	if err != nil {
		return "", exportapp.ErrInvalid
	}
	var state string
	err = storage.inTenant(ctx, claim.Candidate.TenantID,
		func(boundContext context.Context, bound *workerBound) error {
			schema, schemaErr := maintenanceSchema(boundContext, bound, claim)
			if schemaErr != nil {
				return schemaErr
			}
			arguments := maintenanceResolutionArguments(schema, claim, resolution,
				effectReceiptID, effectReceiptSHA)
			return bound.queryRow(boundContext, workerMaintenanceUnknownResolveSQL,
				arguments...).Scan(&state)
		})
	if err != nil {
		return "", mapDispatchError(err)
	}
	terminal := exportapp.State(state)
	if terminal != resolution.TerminalState ||
		!validMaintenanceTerminal(claim.Candidate.Kind, terminal) {
		return "", exportapp.ErrUnavailable
	}
	return terminal, nil
}

func platformMaintenanceCandidate(
	candidate exportapp.MaintenanceUnknownCandidate,
) platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate {
	return platformasync.AnalyticsDashboardExportMaintenanceUnknownCandidate{
		Kind: candidate.Kind, TenantID: candidate.TenantID,
		TenantSchema: candidate.TenantSchema, ExportID: candidate.ExportID,
		AuthorityGroupID:         candidate.AuthorityGroupID,
		AuthorityGroupGeneration: candidate.AuthorityGroupGeneration,
		CutoverGeneration:        candidate.CutoverGeneration, JobID: candidate.JobID,
		AttemptID: candidate.AttemptID, AttemptNumber: candidate.AttemptNumber,
		LeaseGeneration:      candidate.LeaseGeneration,
		OriginalLeaseOwnerID: candidate.OriginalLeaseOwnerID,
		EffectID:             candidate.EffectID,
		EffectRequestSHA256:  candidate.EffectRequestSHA256,
		UnknownAt:            candidate.UnknownAt,
	}
}

func validMaintenanceResolution(claim exportapp.MaintenanceUnknownClaim,
	resolution exportapp.MaintenanceResolution,
) bool {
	evidence := resolution.Evidence
	base := resolution.ReceiptID.Version() == 7 &&
		resolution.SHA256 != [32]byte{} &&
		resolution.TerminalState == claim.TerminalState &&
		validMaintenanceTerminal(claim.Candidate.Kind, resolution.TerminalState) &&
		evidence.LocalStateSHA256 == claim.StoredEvidence.LocalStateSHA256 &&
		evidence.LocalStateAt.Equal(claim.StoredEvidence.LocalStateAt) &&
		!resolution.ResolvedAt.IsZero() &&
		resolution.ResolvedAt.After(claim.Candidate.UnknownAt) &&
		!resolution.ResolvedAt.Before(evidence.LocalStateAt)
	if !base {
		return false
	}
	if resolution.Kind == exportapp.UnknownResolutionAbsent {
		return emptyMaintenanceObjectEvidence(evidence)
	}
	if resolution.Kind != exportapp.UnknownResolutionDeleted ||
		len(evidence.ObjectKey) < 1 || len(evidence.ObjectKey) > 1024 ||
		len(evidence.ObjectVersion) < 1 || len(evidence.ObjectVersion) > 512 ||
		len(evidence.DeleteProviderRequestID) < 1 ||
		len(evidence.DeleteProviderRequestID) > 512 ||
		evidence.DeleteResponseSHA256 == [32]byte{} ||
		evidence.AbsenceSHA256 == [32]byte{} || evidence.DeletedAt.IsZero() ||
		resolution.ResolvedAt.Before(evidence.DeletedAt) {
		return false
	}
	if claim.Mode == exportapp.MaintenanceCleanupArtifact &&
		evidence.DeletedAt.Before(evidence.LocalStateAt) {
		return false
	}
	return validMaintenanceReconcileEvidence(evidence)
}

func emptyMaintenanceObjectEvidence(evidence exportapp.MaintenanceResolutionEvidence) bool {
	return evidence.ObjectKey == "" && evidence.ObjectVersion == "" &&
		evidence.ReconcileProviderRequestID == "" &&
		evidence.DeleteProviderRequestID == "" &&
		evidence.ReconcileEvidenceSHA256 == [32]byte{} &&
		evidence.DeleteResponseSHA256 == [32]byte{} &&
		evidence.AbsenceSHA256 == [32]byte{} && evidence.UploadedAt.IsZero() &&
		evidence.ReconciledAt.IsZero() && evidence.DeletedAt.IsZero()
}

func validMaintenanceReconcileEvidence(
	evidence exportapp.MaintenanceResolutionEvidence,
) bool {
	if evidence.ReconcileProviderRequestID == "" {
		return evidence.ReconcileEvidenceSHA256 == [32]byte{} &&
			evidence.UploadedAt.IsZero() && evidence.ReconciledAt.IsZero()
	}
	return len(evidence.ReconcileProviderRequestID) <= 512 &&
		evidence.ReconcileEvidenceSHA256 != [32]byte{} &&
		!evidence.ReconciledAt.IsZero() &&
		(evidence.UploadedAt.IsZero() ||
			!evidence.ReconciledAt.Before(evidence.UploadedAt)) &&
		!evidence.DeletedAt.Before(evidence.ReconciledAt)
}

func maintenanceResolutionArguments(schema string,
	claim exportapp.MaintenanceUnknownClaim, resolution exportapp.MaintenanceResolution,
	effectReceiptID uuid.UUID, effectReceiptSHA [32]byte,
) []any {
	evidence := resolution.Evidence
	return append(maintenanceClaimArguments(schema, claim),
		resolution.ReceiptID, effectReceiptID, effectReceiptSHA[:], resolution.Kind,
		resolution.SHA256[:], resolution.TerminalState,
		evidence.LocalStateSHA256[:], evidence.LocalStateAt,
		optionalText(evidence.ObjectKey), optionalText(evidence.ObjectVersion),
		optionalText(evidence.ReconcileProviderRequestID),
		optionalDigest(evidence.ReconcileEvidenceSHA256), optionalTime(evidence.UploadedAt),
		optionalTime(evidence.ReconciledAt), optionalText(evidence.DeleteProviderRequestID),
		optionalDigest(evidence.DeleteResponseSHA256), optionalDigest(evidence.AbsenceSHA256),
		optionalTime(evidence.DeletedAt), resolution.ResolvedAt)
}
