package restoreprogressv1

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

const timeFormat = time.RFC3339Nano

var (
	digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	codePattern   = regexp.MustCompile(`^[a-z0-9_]{1,64}$`)
)

func (value Payload) valid() bool {
	occurredAt, err := time.Parse(timeFormat, value.OccurredAt)
	common := err == nil && canonicalTime(occurredAt) &&
		value.Schema == Schema && validUUID(value.EnvironmentID) &&
		validUUID(value.OperationID) && validUUID(value.AttemptID) &&
		value.AttemptID[14] == '7' &&
		digestPattern.MatchString(value.PlanSHA256) &&
		validUUID(value.PreviousDatabaseSystemID) &&
		digestPattern.MatchString(value.PreviousDatabaseLocatorSHA256) &&
		digestPattern.MatchString(value.NewDatabaseLocatorSHA256) &&
		value.PreviousDatabaseLocatorSHA256 != value.NewDatabaseLocatorSHA256
	if !common {
		return false
	}
	switch value.State {
	case StateReadyToStop:
		return value.NewDatabaseSystemID == "" && value.DrainAckSHA256 == "" &&
			value.SwitchAckSHA256 == "" && value.FailureCode == ""
	case StateRestoring:
		return value.NewDatabaseSystemID == "" &&
			digestPattern.MatchString(value.DrainAckSHA256) &&
			value.SwitchAckSHA256 == "" && value.FailureCode == ""
	case StateReadyToSwitch:
		return validUUID(value.NewDatabaseSystemID) &&
			digestPattern.MatchString(value.DrainAckSHA256) &&
			value.SwitchAckSHA256 == "" && value.FailureCode == ""
	case StateCompleted:
		return validUUID(value.NewDatabaseSystemID) &&
			digestPattern.MatchString(value.DrainAckSHA256) &&
			digestPattern.MatchString(value.SwitchAckSHA256) &&
			value.FailureCode == ""
	case StateFailed:
		return codePattern.MatchString(value.FailureCode)
	default:
		return false
	}
}

func validUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%1000 == 0
}
