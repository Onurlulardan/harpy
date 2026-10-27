package postgres

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func validOperationRecord(record ports.OperationRecord) bool {
	if record.OperationID == "" || record.Purpose == "" ||
		record.State.Validate() != nil || len(record.Proposal.Bytes) == 0 ||
		record.Proposal.Value.OperationID != record.OperationID ||
		record.Proposal.Value.Purpose != record.Purpose ||
		record.Proposal.Value.StateSHA256 != record.State.SHA256 ||
		record.Proposal.Value.RegistryBindingID != record.State.Value.RegistryBindingID ||
		record.Proposal.Value.ScopeKind != record.State.Value.ScopeKind ||
		record.Proposal.Value.ScopeID != record.State.Value.ScopeID {
		return false
	}
	_, issuedOK := parseOperationInstant(record.Proposal.Value.IssuedAt)
	_, expiresOK := parseOperationInstant(record.Proposal.Value.ExpiresAt)
	return issuedOK && expiresOK
}

func parseOperationInstant(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339, value)
	return parsed, err == nil && parsed.Location() == time.UTC &&
		parsed.Nanosecond() == 0 && parsed.Format(time.RFC3339) == value
}
