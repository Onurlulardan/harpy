package catalogcomponentv2

import (
	"sort"
	"strconv"

	catalogcomponentv1 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
)

func (value OwnerResult) ValidForLifecycle(command LifecycleCommand) bool {
	if !command.Valid() || value.TenantID != command.TenantID ||
		value.PackageID != command.PackageID || value.OperationID != command.OperationID ||
		value.Component != command.Component || value.ReferenceEpoch != command.ReferenceEpoch ||
		value.AuthoritySHA256 != "" || !validOwnerTarget(value) ||
		!validDigest(value.ResultSHA256) || value.ResultSHA256 != ResultSHA256(value) ||
		!value.Ready || !value.RetryAt.IsZero() || !canonicalTime(value.CompletedAt) ||
		len(value.Blockers) != 0 {
		return false
	}
	return command.Action == "retire" && value.State == StateRetired ||
		command.Action == "restore" && value.State == StateRestored
}

func (value OwnerResult) ValidFor(command PurgeCommand) bool {
	return value.ValidFinalizeFor(command)
}

func (value OwnerResult) ValidPreflightFor(command PurgeCommand) bool {
	if !value.validPurgeBinding(command) {
		return false
	}
	switch value.State {
	case StateReady:
		return value.Ready && value.RetryAt.IsZero() &&
			canonicalTime(value.CompletedAt) && len(value.Blockers) == 0
	case StateWaiting, StateCooling:
		return !value.Ready && canonicalTime(value.RetryAt) &&
			value.CompletedAt.IsZero() && len(value.Blockers) == 0
	case StateBlocked:
		return !value.Ready && value.RetryAt.IsZero() &&
			value.CompletedAt.IsZero() && validBlockers(value.Blockers)
	default:
		return false
	}
}

func (value OwnerResult) ValidBeginFor(command PurgeCommand) bool {
	if !value.validPurgeBinding(command) {
		return false
	}
	switch value.State {
	case StatePurgePending:
		return value.Ready && value.RetryAt.IsZero() &&
			canonicalTime(value.CompletedAt) && len(value.Blockers) == 0
	case StateWaiting, StateCooling:
		return !value.Ready && canonicalTime(value.RetryAt) &&
			value.CompletedAt.IsZero() && len(value.Blockers) == 0
	case StateBlocked:
		return !value.Ready && value.RetryAt.IsZero() &&
			value.CompletedAt.IsZero() && validBlockers(value.Blockers)
	default:
		return false
	}
}

func (value OwnerResult) ValidFinalizeFor(command PurgeCommand) bool {
	if !value.validPurgeBinding(command) {
		return false
	}
	switch value.State {
	case StateWaiting, StatePurgePending:
		return !value.Ready && canonicalTime(value.RetryAt) &&
			value.CompletedAt.IsZero() && len(value.Blockers) == 0
	case StateBlocked:
		return !value.Ready && value.RetryAt.IsZero() &&
			value.CompletedAt.IsZero() && validBlockers(value.Blockers)
	case StatePurged:
		return value.Ready && value.RetryAt.IsZero() &&
			canonicalTime(value.CompletedAt) && len(value.Blockers) == 0
	default:
		return false
	}
}

func (value OwnerResult) validPurgeBinding(command PurgeCommand) bool {
	return command.Valid() && value.TenantID == command.Authority.TenantID &&
		value.PackageID == command.Authority.PackageID &&
		value.OperationID == command.Authority.OperationID &&
		value.Component == command.Component &&
		value.ReferenceEpoch == command.Authority.ReferenceEpoch &&
		value.AuthoritySHA256 == command.Authority.AuthoritySHA256 &&
		validOwnerTarget(value) && validDigest(value.ResultSHA256) &&
		value.ResultSHA256 == ResultSHA256(value)
}

func (command PurgeCommand) Valid() bool {
	return command.Authority.Valid() && validComponent(command.Component) &&
		idempotencyPattern.MatchString(command.IdempotencyKey)
}

func ResultSHA256(value OwnerResult) string {
	fields := []string{
		"catalog.component-owner-result.v2", value.ReceiptID, value.TenantID,
		value.PackageID, value.Component.ID, value.Component.Owner,
		strconv.FormatInt(value.Component.Generation, 10), value.Component.Contract,
		value.Component.ContentSHA256, value.OperationID, value.ResourceID,
		value.VersionID, value.HeadID, strconv.FormatUint(value.ReferenceEpoch, 10),
		value.State, value.AuthoritySHA256, strconv.FormatBool(value.Ready),
		formatTime(value.RetryAt), formatTime(value.CompletedAt),
	}
	blockers := append([]catalogcomponentv1.Blocker(nil), value.Blockers...)
	sort.Slice(blockers, func(i, j int) bool {
		left, right := blockers[i], blockers[j]
		return left.Owner+"\x00"+left.Code+"\x00"+left.Digest <
			right.Owner+"\x00"+right.Code+"\x00"+right.Digest
	})
	for _, blocker := range blockers {
		fields = append(fields, blocker.Owner, blocker.Code, blocker.Digest)
	}
	return digestFields(fields...)
}

func validBlockers(values []catalogcomponentv1.Blocker) bool {
	if len(values) == 0 || len(values) > 100 {
		return false
	}
	for _, value := range values {
		if value.Owner == "" || value.Code == "" || !validDigest(value.Digest) {
			return false
		}
	}
	return true
}

func validOwnerTarget(value OwnerResult) bool {
	return validUUID(value.ReceiptID) &&
		(value.ResourceID == "" && value.VersionID == "" && value.HeadID == "" ||
			validUUID(value.ResourceID) && validUUID(value.VersionID) && validUUID(value.HeadID))
}

func (value OwnerResult) HasResourceBinding() bool {
	return validUUID(value.ResourceID) && validUUID(value.VersionID) && validUUID(value.HeadID)
}
