package provider

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type boundaryFactDocument struct {
	DueAt            string `json:"due_at"`
	ExpectedRevision int64  `json:"expected_revision"`
	JobID            string `json:"job_id"`
	Kind             string `json:"kind"`
	Outcome          string `json:"outcome"`
}

type boundaryFact struct {
	jobID            uuid.UUID
	kind             string
	dueAt            time.Time
	expectedRevision int64
	outcome          string
}

func decodeBoundaryFact(
	document boundaryFactDocument,
	snapshot subscriptionFact,
) (boundaryFact, error) {
	jobID, err := uuid.Parse(document.JobID)
	dueAt, dueErr := decodeInstant(document.DueAt)
	value := boundaryFact{
		jobID: jobID, kind: document.Kind, dueAt: dueAt,
		expectedRevision: document.ExpectedRevision, outcome: document.Outcome,
	}
	if err != nil || dueErr != nil || jobID == uuid.Nil ||
		jobID.Version() != uuid.Version(7) || value.expectedRevision < 1 ||
		!value.validFor(snapshot) {
		return boundaryFact{}, fmt.Errorf("subscription boundary fact is invalid")
	}
	return value, nil
}

func (value boundaryFact) validFor(snapshot subscriptionFact) bool {
	if value.kind != "manual_change" && value.kind != "trial_end" &&
		value.kind != "period_end" {
		return false
	}
	switch value.outcome {
	case "paid":
		return snapshot.state == "active" && snapshot.period != nil &&
			snapshot.period.Start().Equal(value.dueAt) &&
			snapshot.trialEndsAt == nil && snapshot.graceEndsAt == nil
	case "grace":
		return (value.kind == "manual_change" || value.kind == "trial_end") &&
			snapshot.state == "past_due" && snapshot.graceEndsAt != nil &&
			snapshot.graceEndsAt.After(value.dueAt) && snapshot.trialEndsAt == nil
	case "unpaid":
		return (snapshot.state == "cancelled" || snapshot.state == "expired") &&
			snapshot.period == nil && snapshot.trialEndsAt == nil &&
			snapshot.graceEndsAt == nil
	default:
		return false
	}
}
