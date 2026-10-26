package restoreorchestrationv1

import (
	"encoding/base64"
	"regexp"
	"slices"
	"time"

	"github.com/google/uuid"
)

const timeFormat = time.RFC3339Nano

var (
	digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	keyIDPattern  = regexp.MustCompile(`^[a-z][a-z0-9._:-]{0,127}$`)
)

func validTrust(value trustWire) bool {
	public, err := base64.RawURLEncoding.Strict().DecodeString(value.PublicKey)
	return err == nil && len(public) == 32 &&
		value.Schema == TrustSchema && validUUID(value.EnvironmentID) &&
		digestPattern.MatchString(value.DeploymentBindingSHA256) &&
		value.Generation > 0 && keyIDPattern.MatchString(value.KeyID) &&
		value.Algorithm == Algorithm
}

func validAcknowledgement(value acknowledgementWire, trust Trust, now time.Time) bool {
	issuedAt, issuedErr := time.Parse(timeFormat, value.IssuedAt)
	expiresAt, expiryErr := time.Parse(timeFormat, value.ExpiresAt)
	common := value.Schema == AcknowledgementSchema &&
		value.EnvironmentID == trust.EnvironmentID() &&
		validUUID(value.EnvironmentID) && validUUID(value.OperationID) &&
		validUUID(value.AttemptID) && value.AttemptID[14] == '7' &&
		digestPattern.MatchString(value.PlanSHA256) &&
		value.Sequence > 0 && issuedErr == nil && expiryErr == nil &&
		canonicalTime(issuedAt) && canonicalTime(expiresAt) &&
		expiresAt.After(issuedAt) && !now.Before(issuedAt) && now.Before(expiresAt) &&
		value.KeyGeneration == trust.Generation() && value.KeyID == trust.KeyID() &&
		value.Algorithm == Algorithm && validUUID(value.PreviousDatabaseSystemID) &&
		digestPattern.MatchString(value.PreviousDatabaseLocatorSHA256) &&
		digestPattern.MatchString(value.NewDatabaseLocatorSHA256) &&
		value.PreviousDatabaseLocatorSHA256 != value.NewDatabaseLocatorSHA256
	if !common {
		return false
	}
	switch value.Purpose {
	case PurposeDrain:
		return value.Sequence == 1 && value.NewDatabaseSystemID == "" &&
			value.SwitchOutcome == "" &&
			slices.Equal(value.StoppedServices,
				[]string{"core-api", "core-worker", "frontend-v3"})
	case PurposeSwitch:
		return value.Sequence == 2 && validUUID(value.NewDatabaseSystemID) &&
			value.NewDatabaseSystemID != value.PreviousDatabaseSystemID &&
			len(value.StoppedServices) == 0 &&
			value.SwitchOutcome == "switched_compare_and_swap"
	default:
		return false
	}
}

func matchesExpected(value acknowledgementWire, expected Expected) bool {
	if value.Purpose != expected.Purpose ||
		value.EnvironmentID != expected.EnvironmentID ||
		value.OperationID != expected.OperationID ||
		value.AttemptID != expected.AttemptID ||
		value.PlanSHA256 != expected.PlanSHA256 ||
		value.PreviousDatabaseSystemID != expected.PreviousDatabaseSystemID ||
		value.PreviousDatabaseLocatorSHA256 !=
			expected.PreviousDatabaseLocatorSHA256 ||
		value.NewDatabaseLocatorSHA256 != expected.NewDatabaseLocatorSHA256 {
		return false
	}
	return expected.Purpose != PurposeSwitch ||
		value.NewDatabaseSystemID == expected.NewDatabaseSystemID
}

func validUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%1000 == 0
}
