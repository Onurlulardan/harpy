package fullrollout

import (
	"crypto/sha256"
	"encoding/binary"
	"hash"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/cutover/grouprehearsal"
)

func digestCompletion(value PersistedCompletion) [sha256.Size]byte {
	digest := sha256.New()
	writeText(digest, "harpy/full-rollout-completion/v1")
	_, _ = digest.Write(value.ReceiptID[:])
	_, _ = digest.Write(value.EnvironmentID[:])
	writeTime(digest, value.RecordedAt)
	evidence := value.Evidence
	for _, step := range evidence.Steps {
		writeText(digest, string(step.Step.Scope))
		writeText(digest, step.Step.Domain)
		_, _ = digest.Write(step.ScopeID[:])
		_, _ = digest.Write(step.RequestedBy[:])
		_, _ = digest.Write(step.ApprovedBy[:])
		writeInt64(digest, step.Generation)
		writeText(digest, step.State)
		writeText(digest, step.Writer)
		writeDigests(digest, step.AcceptanceSHA256, step.ReverseRehearsalSHA256,
			step.SourceSHA256, step.TargetSHA256)
		writeUint64s(digest, step.SourceCount, step.TargetCount, step.PendingJobs,
			step.PendingOutbox, step.PendingEffects)
	}
	for _, cohort := range evidence.Cohorts {
		writeText(digest, string(cohort.Cohort))
		writeTime(digest, cohort.StartedAt)
		writeTime(digest, cohort.EndedAt)
		writeUint64s(digest, cohort.EligibleRequests, cohort.ComparedRequests,
			cohort.MismatchCount)
		writeMetrics(digest, cohort.Metrics)
	}
	writeTime(digest, evidence.Now)
	writeTime(digest, evidence.LastLegacyHitAt)
	writeUint64s(digest, evidence.LegacyRouteHits, evidence.V1Writers,
		evidence.V1Consumers, evidence.V1QueueMessages, evidence.V1Credentials)
	writeDigests(digest, evidence.ShadowReadSHA256, evidence.LogicalDeltaSHA256,
		evidence.ReconcileSHA256, evidence.ProductionSnapshotSHA256,
		evidence.V1RemovalSHA256, evidence.BackupRestoreSHA256,
		evidence.FrontendContractSHA256, evidence.MCPContractSHA256,
		evidence.RollbackWindowClosureSHA256)
	writeTime(digest, evidence.RollbackWindowClosedAt)
	var result [sha256.Size]byte
	copy(result[:], digest.Sum(nil))
	return result
}

func writeMetrics(target hash.Hash, value grouprehearsal.Metrics) {
	for _, metric := range []int64{
		value.V2ErrorBasisPoints, value.BaselineErrorBasisPoints,
		value.V2P95Milliseconds, value.BaselineP95Milliseconds,
		value.AuthSuccessBasisPoints, value.BaselineAuthSuccessBasisPoints,
		value.OldestDueJobLagSeconds, value.BaselineJobLagSeconds,
		value.DLQBasisPoints,
	} {
		writeInt64(target, metric)
	}
	writeUint64s(target, value.TenantIsolationIncidents, value.AuthorizationBypasses,
		value.CredentialLeaks, value.DataLossIncidents, value.PermanentSecurityJobs)
}

func writeText(target hash.Hash, value string) {
	writeUint64(target, uint64(len(value)))
	_, _ = target.Write([]byte(value))
}

func writeTime(target hash.Hash, value time.Time) {
	writeInt64(target, value.UnixMicro())
}

func writeInt64(target hash.Hash, value int64) { writeUint64(target, uint64(value)) }

func writeUint64(target hash.Hash, value uint64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	_, _ = target.Write(encoded[:])
}

func writeUint64s(target hash.Hash, values ...uint64) {
	for _, value := range values {
		writeUint64(target, value)
	}
}

func writeDigests(target hash.Hash, values ...[sha256.Size]byte) {
	for _, value := range values {
		_, _ = target.Write(value[:])
	}
}
