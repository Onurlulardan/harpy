package platformlicense

import (
	"encoding/hex"
	"regexp"

	"github.com/google/uuid"
)

var idempotencyPattern = regexp.MustCompile(
	`^[A-Za-z0-9][A-Za-z0-9._:-]{15,127}$`,
)

func validEnvironment(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}

func validDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32
}

func validLicense(value License, environmentID uuid.UUID) bool {
	return value.LicenseID != uuid.Nil && value.EnvironmentID == environmentID &&
		value.SnapshotSequence > 0 && validDigest(value.SnapshotSHA256) &&
		(value.State == "valid" || value.State == "grace" ||
			value.State == "build_off") &&
		value.ReleaseChannel != "" && len(value.Features) > 0 &&
		len(value.Limits) > 0 && !value.DatabaseNow.IsZero()
}

func validOperation(value RefreshOperation, operationID uuid.UUID) bool {
	if value.OperationID != operationID || value.OperationID.Version() != uuid.Version(7) ||
		value.CreatedAt.IsZero() {
		return false
	}
	switch value.State {
	case "pending", "leased", "accepted", "failed", "fenced":
		return true
	default:
		return false
	}
}
