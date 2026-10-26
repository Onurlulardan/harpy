package lifecycle

import (
	"time"

	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	"github.com/google/uuid"
)

type operationRecord struct {
	ID, InstallationID, PackageID uuid.UUID
	Kind, State                   string
	Plan                          []byte
	PlanSHA256                    string
	RetryOf                       uuid.NullUUID
	StartedAt                     time.Time
	DestructiveStartedAt          *time.Time
	FailureCode                   *string
	Revision                      int64
}

func scanOperation(source row) (catalogdomain.Operation, uuid.UUID, error) {
	var record operationRecord
	err := source.Scan(&record.ID, &record.InstallationID, &record.PackageID,
		&record.Kind, &record.State, &record.Plan, &record.PlanSHA256, &record.RetryOf,
		&record.StartedAt, &record.DestructiveStartedAt, &record.FailureCode, &record.Revision)
	if err != nil {
		return catalogdomain.Operation{}, uuid.Nil, err
	}
	record.normalizeTimes()
	plan, err := decodePlan(record.Plan, record.PlanSHA256)
	if err != nil || plan.PackageID.String() != record.PackageID.String() || record.Revision < 1 {
		return catalogdomain.Operation{}, uuid.Nil, unavailable()
	}
	id, err := packagecontract.ParseStableID(record.ID.String())
	if err != nil {
		return catalogdomain.Operation{}, uuid.Nil, unavailable()
	}
	operation, err := catalogdomain.NewOperation(id, catalogdomain.OperationKind(record.Kind), plan, record.StartedAt)
	if err != nil || !validOperationState(catalogdomain.OperationState(record.State)) {
		return catalogdomain.Operation{}, uuid.Nil, unavailable()
	}
	operation.State, operation.Revision = catalogdomain.OperationState(record.State), uint64(record.Revision)
	if record.FailureCode != nil {
		operation.FailureCode = *record.FailureCode
	}
	if record.RetryOf.Valid {
		retryID, parseErr := packagecontract.ParseStableID(record.RetryOf.UUID.String())
		if parseErr != nil || retryID == id {
			return catalogdomain.Operation{}, uuid.Nil, unavailable()
		}
		operation.RetryOf = &retryID
	}
	if record.DestructiveStartedAt != nil {
		if !canonicalTime(*record.DestructiveStartedAt) || operation.Kind != catalogdomain.OperationPurge {
			return catalogdomain.Operation{}, uuid.Nil, unavailable()
		}
		operation.DestructiveStarted = record.DestructiveStartedAt
	}
	return operation, record.InstallationID, nil
}

func (record *operationRecord) normalizeTimes() {
	record.StartedAt = record.StartedAt.UTC()
	if record.DestructiveStartedAt != nil {
		destructiveStarted := record.DestructiveStartedAt.UTC()
		record.DestructiveStartedAt = &destructiveStarted
	}
}

func validOperationState(value catalogdomain.OperationState) bool {
	switch value {
	case catalogdomain.OperationPlanned, catalogdomain.OperationApplying,
		catalogdomain.OperationWaiting, catalogdomain.OperationVerifying,
		catalogdomain.OperationActivating, catalogdomain.OperationCancelling,
		catalogdomain.OperationSucceeded, catalogdomain.OperationBlocked,
		catalogdomain.OperationFailed, catalogdomain.OperationCancelled:
		return true
	default:
		return false
	}
}

func canonicalTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1000 == 0
}
