package cutoverv1

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const terminalMaximumSafeInteger = int64(9007199254740991)

var terminalSnapshotNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,255}$`)

func ValidateTerminalEvidence(value TerminalEvidence) error {
	if value.ContractVersion != "cutover-terminal-evidence.v1" ||
		!validUUID(value.OperationID) || !validUUID(value.RunID) ||
		value.OperationID == value.RunID || !validDigest(value.LeaseManifestSHA256) ||
		!validDigest(value.PlanSHA256) || !validDigest(value.TransformManifestSHA256) {
		return contractError("terminal evidence identity")
	}
	switch value.TerminalState {
	case "completed":
		return validateCompletedTerminalEvidence(value)
	case "failed", "cancelled":
		return validateAbortTerminalEvidence(value)
	default:
		return contractError("terminal evidence state")
	}
}

func validateCompletedTerminalEvidence(value TerminalEvidence) error {
	if value.FinalLSN == nil || value.InitialCopy == nil || value.CheckpointCount == nil ||
		value.JournalTransactionCount == nil || value.AppliedChangeCount == nil ||
		value.JournalCoverageRootSHA256 == nil || value.OutcomeCode != nil ||
		value.OutcomeEvidenceSHA256 != nil || !validTerminalLSN(*value.FinalLSN) {
		return contractError("completed terminal evidence shape")
	}
	if err := validateTerminalInitialCopy(*value.InitialCopy, value); err != nil {
		return err
	}
	if *value.CheckpointCount != len(value.Checkpoints) ||
		*value.CheckpointCount != value.InitialCopy.PartitionCount ||
		*value.CheckpointCount < 1 || *value.CheckpointCount > 10000 {
		return contractError("terminal checkpoint count")
	}
	for index, checkpoint := range value.Checkpoints {
		if err := validateTerminalCheckpoint(checkpoint, *value.FinalLSN); err != nil {
			return err
		}
		if index > 0 && !terminalCheckpointLess(value.Checkpoints[index-1], checkpoint) {
			return contractError("terminal checkpoint order")
		}
	}
	if !terminalRelationOrdinalsComplete(value.Checkpoints, value.InitialCopy.RelationCount) {
		return contractError("terminal checkpoint relation set")
	}
	return validateTerminalJournalCoverage(value)
}

func validateAbortTerminalEvidence(value TerminalEvidence) error {
	if value.FinalLSN != nil || value.InitialCopy != nil || value.CheckpointCount != nil ||
		value.Checkpoints != nil || value.JournalTransactionCount != nil ||
		value.AppliedChangeCount != nil || value.JournalCoverageRootSHA256 != nil ||
		value.OutcomeCode == nil || value.OutcomeEvidenceSHA256 == nil ||
		!validDigest(*value.OutcomeEvidenceSHA256) {
		return contractError("abort terminal evidence shape")
	}
	code := *value.OutcomeCode
	if value.TerminalState == "failed" && code != "source_failed" &&
		code != "target_failed" && code != "verification_failed" && code != "lease_expired" {
		return contractError("failed terminal outcome")
	}
	if value.TerminalState == "cancelled" &&
		code != "operator_cancelled" && code != "superseded" {
		return contractError("cancelled terminal outcome")
	}
	return nil
}

func validateTerminalJournalCoverage(value TerminalEvidence) error {
	journalCount, appliedCount := *value.JournalTransactionCount, *value.AppliedChangeCount
	if journalCount < 0 || appliedCount < 0 || journalCount > terminalMaximumSafeInteger ||
		appliedCount > terminalMaximumSafeInteger ||
		!validDigest(*value.JournalCoverageRootSHA256) ||
		(journalCount == 0) != (appliedCount == 0) {
		return contractError("terminal journal coverage")
	}
	if journalCount == 0 && *value.JournalCoverageRootSHA256 !=
		digestTerminalLeaves(terminalCoverageRootDomain, nil) {
		return contractError("empty terminal journal coverage")
	}
	return nil
}

func validateTerminalInitialCopy(initial TerminalInitialCopyEvidence, value TerminalEvidence) error {
	if initial.ManifestSHA256 != value.LeaseManifestSHA256 ||
		!terminalSnapshotNamePattern.MatchString(initial.SnapshotName) ||
		!validTerminalLSN(initial.SnapshotLSN) ||
		compareTerminalLSN(initial.SnapshotLSN, *value.FinalLSN) > 0 ||
		initial.RelationCount < 1 || initial.RelationCount > 4096 ||
		initial.PartitionCount < initial.RelationCount || initial.PartitionCount > 10000 ||
		initial.RowCount < 0 || initial.RowCount > terminalMaximumSafeInteger ||
		!validDigest(initial.RootSHA256) {
		return contractError("terminal initial copy")
	}
	return nil
}

func validateTerminalCheckpoint(value TerminalCheckpointEvidence, finalLSN string) error {
	if value.RelationOrdinal < 1 || value.RelationOrdinal > 4096 ||
		len(value.CheckpointKey) < 1 || len(value.CheckpointKey) > 512 ||
		!utf8.ValidString(value.CheckpointKey) || containsControl(value.CheckpointKey) ||
		!validTerminalLSN(value.AppliedLSN) ||
		compareTerminalLSN(value.AppliedLSN, finalLSN) < 0 || value.AppliedRows < 0 ||
		value.AppliedRows > terminalMaximumSafeInteger || value.BatchCount < 1 ||
		int64(value.BatchCount) > 2147483647 || !validDigest(value.BatchRootSHA256) {
		return contractError("terminal checkpoint")
	}
	return nil
}

func terminalCheckpointLess(left, right TerminalCheckpointEvidence) bool {
	return left.RelationOrdinal < right.RelationOrdinal ||
		left.RelationOrdinal == right.RelationOrdinal && left.CheckpointKey < right.CheckpointKey
}

func terminalRelationOrdinalsComplete(values []TerminalCheckpointEvidence, count int) bool {
	expected := 1
	for _, value := range values {
		if value.RelationOrdinal == expected {
			expected++
		} else if value.RelationOrdinal != expected-1 {
			return false
		}
	}
	return expected == count+1
}

func validTerminalLSN(value string) bool {
	if !validLSN(value) {
		return false
	}
	high, low, ok := parseTerminalLSN(value)
	return ok && value == strings.ToUpper(strconv.FormatUint(high, 16))+"/"+
		strings.ToUpper(strconv.FormatUint(low, 16))
}

func parseTerminalLSN(value string) (uint64, uint64, bool) {
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return 0, 0, false
	}
	high, highErr := strconv.ParseUint(parts[0], 16, 32)
	low, lowErr := strconv.ParseUint(parts[1], 16, 32)
	return high, low, highErr == nil && lowErr == nil
}

func compareTerminalLSN(left, right string) int {
	leftHigh, leftLow, _ := parseTerminalLSN(left)
	rightHigh, rightLow, _ := parseTerminalLSN(right)
	if leftHigh != rightHigh {
		if leftHigh < rightHigh {
			return -1
		}
		return 1
	}
	if leftLow < rightLow {
		return -1
	}
	if leftLow > rightLow {
		return 1
	}
	return 0
}
