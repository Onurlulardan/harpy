package cutoverv1

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
)

var lsnPattern = regexp.MustCompile(`^(?:0|[1-9A-F][0-9A-F]*)/[0-9A-F]{1,8}$`)

func ValidateChangeBatch(batch ChangeBatch) error {
	if batch.ContractVersion != "cutover-change-batch.v1" || !validDigest(batch.BatchSHA256) {
		return contractError("change batch contract")
	}
	payload := batch.Payload
	if !validUUID(payload.OperationID) || !validUUID(payload.RunID) ||
		!validUUID(payload.ScopeID) || payload.OperationID == payload.RunID ||
		(payload.ScopeKind != "global" && payload.ScopeKind != "tenant") {
		return contractError("change batch identity")
	}
	if payload.RelationOrdinal < 1 || payload.RelationOrdinal > 4096 ||
		len(payload.CheckpointKey) < 1 || len(payload.CheckpointKey) > 512 ||
		containsControl(payload.CheckpointKey) {
		return contractError("change batch checkpoint")
	}
	if payload.CheckpointState != "initial_copy" && payload.CheckpointState != "delta" &&
		payload.CheckpointState != "complete" {
		return contractError("change batch checkpoint state")
	}
	if err := validateSourcePosition(payload.Position, payload.CheckpointState); err != nil {
		return err
	}
	if !validJSONObjectOrNull(payload.LastKey, true) ||
		len(payload.Changes) > MaximumChangeBatchChanges {
		return contractError("change batch payload")
	}
	if len(payload.Changes) == 0 {
		if payload.CheckpointState != "complete" || payload.Position.FinalLSN == nil ||
			payload.Position.CommitLSN != *payload.Position.FinalLSN {
			return contractError("change batch completion barrier")
		}
		return nil
	}
	for index, change := range payload.Changes {
		if err := validateChange(change, index+1); err != nil {
			return err
		}
		if _, err := PersistedChangeOrdinal(payload.Position.BatchOrdinal, change.Ordinal); err != nil {
			return err
		}
	}
	return nil
}

func validateSourcePosition(position SourcePosition, state string) error {
	if len(position.SnapshotName) < 1 || len(position.SnapshotName) > 256 ||
		containsControl(position.SnapshotName) || !validLSN(position.SnapshotLSN) ||
		!validLSN(position.ReceivedLSN) || !validLSN(position.CommitLSN) ||
		position.BatchOrdinal < 1 || position.BatchOrdinal > MaximumChangeBatchOrdinal ||
		!validPositiveDecimal(position.SourceXID) {
		return contractError("change batch source position")
	}
	if state == "complete" {
		if position.FinalLSN == nil || !validLSN(*position.FinalLSN) {
			return contractError("complete batch final position")
		}
	} else if position.FinalLSN != nil {
		return contractError("non-complete batch final position")
	}
	return nil
}

func validateChange(change Change, ordinal int) error {
	if change.Ordinal != ordinal || !validJSONObjectOrNull(change.Key, false) {
		return contractError("change ordinal or key")
	}
	before := validJSONObjectOrNull(change.Before, true)
	after := validJSONObjectOrNull(change.After, true)
	if !before || !after {
		return contractError("change value")
	}
	switch change.Kind {
	case "insert":
		if !isJSONNull(change.Before) && len(change.Before) != 0 || isJSONNull(change.After) {
			return contractError("insert change shape")
		}
	case "update":
		if isJSONNull(change.Before) || isJSONNull(change.After) {
			return contractError("update change shape")
		}
	case "delete":
		if isJSONNull(change.Before) || !isJSONNull(change.After) && len(change.After) != 0 {
			return contractError("delete change shape")
		}
	default:
		return contractError("change kind")
	}
	return nil
}

func validJSONObjectOrNull(value json.RawMessage, nullable bool) bool {
	trimmed := bytes.TrimSpace(value)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return nullable
	}
	var object map[string]json.RawMessage
	return json.Unmarshal(trimmed, &object) == nil && object != nil && len(object) > 0
}

func isJSONNull(value json.RawMessage) bool {
	trimmed := bytes.TrimSpace(value)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

func validLSN(value string) bool { return lsnPattern.MatchString(value) }

func validPositiveDecimal(value string) bool {
	if value == "" || value[0] == '0' || len(value) > 20 {
		return false
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	return err == nil && parsed > 0
}

func ValidateApplyResult(result ApplyResult) error {
	if result.AppliedCount < 0 || !validDigest(result.KeyStateSHA256) ||
		!validDigest(result.StateSHA256) {
		return contractError("apply result")
	}
	return nil
}
