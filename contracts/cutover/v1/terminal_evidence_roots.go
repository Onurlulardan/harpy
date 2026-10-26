package cutoverv1

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"unicode/utf8"
)

const (
	terminalBatchRootDomain    = "cutover-batch-root.v1\n"
	terminalChangesRootDomain  = "cutover-journal-changes-root.v1\n"
	terminalCoverageRootDomain = "cutover-journal-coverage-root.v1\n"
)

func TerminalBatchRootSHA256(leaves []TerminalBatchLeaf) (string, error) {
	if len(leaves) == 0 {
		return "", contractError("terminal batch leaf set")
	}
	framed := make([][]byte, len(leaves))
	priorOrdinal := 0
	for index, leaf := range leaves {
		if leaf.BatchOrdinal < 1 || leaf.BatchOrdinal > 2147483647 ||
			index > 0 && leaf.BatchOrdinal <= priorOrdinal || !validPositiveDecimal(leaf.SourceXID) ||
			!validTerminalLSN(leaf.SourceCommitLSN) || !validDigest(leaf.BatchSHA256) ||
			leaf.AppliedCount < 0 || leaf.AppliedCount > terminalMaximumSafeInteger ||
			!validDigest(leaf.KeyStateSHA256) ||
			!validDigest(leaf.StateSHA256) {
			return "", contractError("terminal batch leaf")
		}
		priorOrdinal = leaf.BatchOrdinal
		framed[index] = terminalLeaf(
			strconv.Itoa(leaf.BatchOrdinal), leaf.SourceXID, leaf.SourceCommitLSN,
			leaf.BatchSHA256, strconv.FormatInt(leaf.AppliedCount, 10),
			leaf.KeyStateSHA256, leaf.StateSHA256,
		)
	}
	return digestTerminalLeaves(terminalBatchRootDomain, framed), nil
}

func TerminalJournalCoverageRootSHA256(
	leaves []TerminalJournalCoverageLeaf,
) (string, int64, error) {
	framed := make([][]byte, len(leaves))
	var changeCount int64
	var priorXID uint64
	for index, leaf := range leaves {
		xid, err := strconv.ParseUint(leaf.SourceXID, 10, 64)
		if err != nil || !validPositiveDecimal(leaf.SourceXID) ||
			xid == 0 || index > 0 && xid <= priorXID ||
			!validTerminalLSN(leaf.SourceCommitLSN) || len(leaf.Changes) == 0 {
			return "", 0, contractError("terminal journal leaf")
		}
		changes, err := terminalChangesRoot(leaf.Changes)
		if err != nil {
			return "", 0, err
		}
		priorXID = xid
		changeCount += int64(len(leaf.Changes))
		framed[index] = terminalLeaf(
			leaf.SourceXID, leaf.SourceCommitLSN,
			strconv.Itoa(len(leaf.Changes)), changes,
		)
	}
	return digestTerminalLeaves(terminalCoverageRootDomain, framed), changeCount, nil
}

func terminalChangesRoot(leaves []TerminalChangeLeaf) (string, error) {
	framed := make([][]byte, len(leaves))
	for index, leaf := range leaves {
		if leaf.RelationOrdinal < 1 || leaf.RelationOrdinal > 4096 ||
			leaf.ChangeOrdinal < 1 || int64(leaf.ChangeOrdinal) > 2147483647 ||
			len(leaf.CheckpointKey) < 1 ||
			len(leaf.CheckpointKey) > 512 || !utf8.ValidString(leaf.CheckpointKey) ||
			containsControl(leaf.CheckpointKey) ||
			(leaf.MutationKind != "insert" && leaf.MutationKind != "update" &&
				leaf.MutationKind != "delete") || !validDigest(leaf.MutationSHA256) {
			return "", contractError("terminal change leaf")
		}
		if index > 0 && !terminalChangeLess(leaves[index-1], leaf) {
			return "", contractError("terminal change leaf order")
		}
		framed[index] = terminalLeaf(strconv.Itoa(leaf.RelationOrdinal),
			strconv.Itoa(leaf.ChangeOrdinal), leaf.CheckpointKey,
			leaf.MutationKind, leaf.MutationSHA256)
	}
	return digestTerminalLeaves(terminalChangesRootDomain, framed), nil
}

func terminalChangeLess(left, right TerminalChangeLeaf) bool {
	return left.RelationOrdinal < right.RelationOrdinal ||
		left.RelationOrdinal == right.RelationOrdinal &&
			(left.ChangeOrdinal < right.ChangeOrdinal ||
				left.ChangeOrdinal == right.ChangeOrdinal && left.CheckpointKey < right.CheckpointKey)
}

func terminalLeaf(values ...string) []byte {
	var buffer bytes.Buffer
	for _, value := range values {
		buffer.WriteString(strconv.Itoa(len([]byte(value))))
		buffer.WriteByte(':')
		buffer.WriteString(value)
	}
	return buffer.Bytes()
}

func digestTerminalLeaves(domain string, leaves [][]byte) string {
	hasher := sha256.New()
	_, _ = hasher.Write([]byte(domain))
	for _, leaf := range leaves {
		_, _ = fmt.Fprintf(hasher, "%d:", len(leaf))
		_, _ = hasher.Write(leaf)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}
