package postgres

import (
	"strings"
	"time"
	"unicode/utf8"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

type StaleJobDispositionCommand struct {
	jobID       uuid.UUID
	decision    jobsv1.StaleJobDecision
	reason      string
	ticket      string
	actorID     uuid.UUID
	causationID uuid.UUID
}

func NewStaleJobDispositionCommand(
	jobID uuid.UUID,
	decision jobsv1.StaleJobDecision,
	reason string,
	ticket string,
	actorID uuid.UUID,
	causationID uuid.UUID,
) (StaleJobDispositionCommand, error) {
	command := StaleJobDispositionCommand{
		jobID: jobID, decision: decision, reason: reason, ticket: ticket,
		actorID: actorID, causationID: causationID,
	}
	if !command.valid() {
		return StaleJobDispositionCommand{}, ErrStaleJobDispositionInvalid
	}
	return command, nil
}

func (command StaleJobDispositionCommand) valid() bool {
	decisionValid := command.decision == jobsv1.StaleJobAdopt ||
		command.decision == jobsv1.StaleJobCancel
	return command.jobID != uuid.Nil && decisionValid && command.actorID != uuid.Nil &&
		command.causationID != uuid.Nil && validDispositionText(
		command.reason, jobsv1.MaximumReasonBytes,
	) && validDispositionText(command.ticket, jobsv1.MaximumTicketBytes)
}

func validDispositionText(value string, maximum int) bool {
	return value != "" && utf8.ValidString(value) && len(value) <= maximum &&
		strings.TrimSpace(value) == value
}

type StaleJobDispositionResult struct {
	completedAt time.Time
	actionID    uuid.UUID
	resultJobID uuid.UUID
	resultRoot  uuid.UUID
	created     bool
}

func (result StaleJobDispositionResult) CompletedAt() time.Time { return result.completedAt }
func (result StaleJobDispositionResult) ActionID() uuid.UUID    { return result.actionID }
func (result StaleJobDispositionResult) Created() bool          { return result.created }
func (result StaleJobDispositionResult) ResultJobID() (uuid.UUID, bool) {
	return result.resultJobID, result.resultJobID != uuid.Nil
}
func (result StaleJobDispositionResult) ResultRootID() (uuid.UUID, bool) {
	return result.resultRoot, result.resultRoot != uuid.Nil
}
