package main

import (
	"encoding/hex"
	"fmt"
	"time"

	corecutover "github.com/Onurlulardan/harpy/core/public/cutover"
)

func (document fullRolloutDocument) evidence() (corecutover.FullRolloutEvidence, error) {
	steps, err := decodeFullRolloutSteps(document.Steps)
	if err != nil {
		return corecutover.FullRolloutEvidence{}, err
	}
	cohorts, err := decodeFullRolloutCohorts(document.Cohorts)
	if err != nil {
		return corecutover.FullRolloutEvidence{}, err
	}
	now, err := parseFullRolloutTime(document.Now)
	if err != nil {
		return corecutover.FullRolloutEvidence{}, err
	}
	lastHit, err := parseFullRolloutTime(document.LastLegacyHitAt)
	if err != nil {
		return corecutover.FullRolloutEvidence{}, err
	}
	closedAt, err := parseFullRolloutTime(document.RollbackWindowClosedAt)
	if err != nil {
		return corecutover.FullRolloutEvidence{}, err
	}
	digests, err := decodeFullRolloutDigests([]string{
		document.ShadowReadSHA256, document.LogicalDeltaSHA256,
		document.ReconcileSHA256, document.ProductionSnapshotSHA256,
		document.V1RemovalSHA256, document.BackupRestoreSHA256,
		document.FrontendContractSHA256, document.RollbackClosureSHA256,
	})
	if err != nil {
		return corecutover.FullRolloutEvidence{}, err
	}
	return corecutover.FullRolloutEvidence{Steps: steps, Cohorts: cohorts, Now: now,
		LastLegacyHitAt: lastHit, LegacyRouteHits: document.LegacyRouteHits,
		V1Writers: document.V1Writers, V1Consumers: document.V1Consumers,
		V1QueueMessages: document.V1QueueMessages, V1Credentials: document.V1Credentials,
		ShadowReadSHA256: digests[0], LogicalDeltaSHA256: digests[1],
		ReconcileSHA256: digests[2], ProductionSnapshotSHA256: digests[3],
		V1RemovalSHA256: digests[4], BackupRestoreSHA256: digests[5],
		FrontendContractSHA256: digests[6], RollbackWindowClosureSHA256: digests[7],
		RollbackWindowClosedAt: closedAt}, nil
}

func parseFullRolloutTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil || parsed.Location() != time.UTC ||
		parsed.Nanosecond()%1_000 != 0 || parsed.Format(time.RFC3339Nano) != value {
		return time.Time{}, fmt.Errorf("completion evidence timestamp is invalid")
	}
	return parsed, nil
}

func parseFullRolloutDigest(value string) ([32]byte, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 || hex.EncodeToString(decoded) != value {
		return [32]byte{}, fmt.Errorf("completion evidence digest is invalid")
	}
	var result [32]byte
	copy(result[:], decoded)
	if result == [32]byte{} {
		return [32]byte{}, fmt.Errorf("completion evidence digest is empty")
	}
	return result, nil
}

func decodeFullRolloutDigests(values []string) ([][32]byte, error) {
	result := make([][32]byte, len(values))
	for index, value := range values {
		decoded, err := parseFullRolloutDigest(value)
		if err != nil {
			return nil, err
		}
		result[index] = decoded
	}
	return result, nil
}

func decodeFullRolloutMetrics(value fullRolloutMetrics) corecutover.FullRolloutMetrics {
	return corecutover.FullRolloutMetrics{V2ErrorBasisPoints: value.V2ErrorBasisPoints,
		BaselineErrorBasisPoints:       value.BaselineErrorBasisPoints,
		V2P95Milliseconds:              value.V2P95Milliseconds,
		BaselineP95Milliseconds:        value.BaselineP95Milliseconds,
		AuthSuccessBasisPoints:         value.AuthSuccessBasisPoints,
		BaselineAuthSuccessBasisPoints: value.BaselineAuthSuccessBasisPoints,
		OldestDueJobLagSeconds:         value.OldestDueJobLagSeconds,
		BaselineJobLagSeconds:          value.BaselineJobLagSeconds,
		DLQBasisPoints:                 value.DLQBasisPoints,
		TenantIsolationIncidents:       value.TenantIsolationIncidents,
		AuthorizationBypasses:          value.AuthorizationBypasses,
		CredentialLeaks:                value.CredentialLeaks, DataLossIncidents: value.DataLossIncidents,
		PermanentSecurityJobs: value.PermanentSecurityJobs}
}
