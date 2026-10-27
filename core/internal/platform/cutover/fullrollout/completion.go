package fullrollout

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/grouprehearsal"
)

const LegacySilenceWindow = 30 * 24 * time.Hour

type CompletionEvidence struct {
	Steps                          []StepEvidence
	Cohorts                        []grouprehearsal.SoakEvidence
	Now, LastLegacyHitAt           time.Time
	LegacyRouteHits                uint64
	V1Writers, V1Consumers         uint64
	V1QueueMessages, V1Credentials uint64
	ShadowReadSHA256               [32]byte
	LogicalDeltaSHA256             [32]byte
	ReconcileSHA256                [32]byte
	ProductionSnapshotSHA256       [32]byte
	V1RemovalSHA256                [32]byte
	BackupRestoreSHA256            [32]byte
	FrontendContractSHA256         [32]byte
	MCPContractSHA256              [32]byte
	RollbackWindowClosureSHA256    [32]byte
	RollbackWindowClosedAt         time.Time
}

func EvaluateCompletion(value CompletionEvidence) error {
	if _, err := validateSteps(value.Steps); err != nil {
		return err
	}
	order := grouprehearsal.CohortOrder()
	if len(value.Cohorts) != len(order) || !validEvidenceTime(value.Now) ||
		!validEvidenceTime(value.LastLegacyHitAt) || value.LastLegacyHitAt.After(value.Now) ||
		value.Now.Sub(value.LastLegacyHitAt) < LegacySilenceWindow ||
		value.LegacyRouteHits != 0 || value.V1Writers != 0 || value.V1Consumers != 0 ||
		value.V1QueueMessages != 0 || value.V1Credentials != 0 ||
		value.ShadowReadSHA256 == [32]byte{} || value.LogicalDeltaSHA256 == [32]byte{} ||
		value.ReconcileSHA256 == [32]byte{} || value.ProductionSnapshotSHA256 == [32]byte{} ||
		value.V1RemovalSHA256 == [32]byte{} ||
		value.BackupRestoreSHA256 == [32]byte{} ||
		value.FrontendContractSHA256 == [32]byte{} || value.MCPContractSHA256 == [32]byte{} ||
		value.RollbackWindowClosureSHA256 == [32]byte{} ||
		!validEvidenceTime(value.RollbackWindowClosedAt) ||
		value.RollbackWindowClosedAt.After(value.Now) ||
		value.RollbackWindowClosedAt.Before(value.LastLegacyHitAt) {
		return ErrInvalidEvidence
	}
	var previousEnd time.Time
	for index, soak := range value.Cohorts {
		if soak.Cohort != order[index] || !validEvidenceTime(soak.StartedAt) ||
			!validEvidenceTime(soak.EndedAt) || soak.EndedAt.After(value.Now) ||
			index > 0 && soak.StartedAt.Before(previousEnd) ||
			grouprehearsal.EvaluateSoak(soak) != nil {
			return ErrInvalidEvidence
		}
		previousEnd = soak.EndedAt
	}
	if value.RollbackWindowClosedAt.Before(previousEnd) {
		return ErrInvalidEvidence
	}
	return nil
}
