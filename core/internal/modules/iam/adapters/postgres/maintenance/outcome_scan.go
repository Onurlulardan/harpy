package maintenance

import (
	"encoding/hex"
	"fmt"
	"time"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	"github.com/google/uuid"
)

type rowScanner interface{ Scan(...any) error }

type nullableOutcome struct {
	jobID     *uuid.UUID
	runAt     *time.Time
	reference *string
	digest    []byte
	dedupe    *string
	priority  *int32
}

func scanOutcome(row rowScanner, kind maintenanceapp.Kind) (maintenanceapp.Outcome, error) {
	var value nullableOutcome
	if err := row.Scan(
		&value.jobID, &value.runAt, &value.reference,
		&value.digest, &value.dedupe, &value.priority,
	); err != nil {
		return maintenanceapp.Outcome{}, err
	}
	return value.outcome(kind)
}

func scanReceiptOutcome(
	row rowScanner,
	kind maintenanceapp.Kind,
) (maintenanceapp.Outcome, bool, error) {
	var found bool
	var value nullableOutcome
	if err := row.Scan(
		&found, &value.jobID, &value.runAt, &value.reference,
		&value.digest, &value.dedupe, &value.priority,
	); err != nil {
		return maintenanceapp.Outcome{}, false, err
	}
	if !found {
		return maintenanceapp.Outcome{}, false, nil
	}
	outcome, err := value.outcome(kind)
	return outcome, true, err
}

func (value nullableOutcome) outcome(
	kind maintenanceapp.Kind,
) (maintenanceapp.Outcome, error) {
	if value.jobID == nil {
		outcome := maintenanceapp.Outcome{}
		return outcome, outcome.Validate(kind)
	}
	if value.runAt == nil || value.reference == nil || len(value.digest) != 32 ||
		value.dedupe == nil || value.priority == nil {
		return maintenanceapp.Outcome{}, maintenanceapp.ErrOutcomeInvalid
	}
	outcome := maintenanceapp.Outcome{Successor: &maintenanceapp.Successor{
		JobID: *value.jobID, RunAt: value.runAt.UTC(),
		PayloadReference: *value.reference,
		PayloadSHA256:    hex.EncodeToString(value.digest),
		DedupeKey:        *value.dedupe, Priority: *value.priority,
	}}
	if err := outcome.Validate(kind); err != nil {
		return maintenanceapp.Outcome{}, fmt.Errorf("scan maintenance outcome: %w", err)
	}
	return outcome, nil
}
