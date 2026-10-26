package async

import (
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

var databaseJobDigest = regexp.MustCompile(`^[0-9a-f]{64}$`)

type DatabaseJobSuccessorInput struct {
	JobID            uuid.UUID
	RunAt            time.Time
	PayloadReference string
	PayloadSHA256    string
	DedupeKey        string
	Priority         int32
}

type DatabaseJobOutcome struct {
	successor     DatabaseJobSuccessorInput
	hasNext       bool
	children      []DatabaseJobChildInput
	cancellations []DatabaseJobCancellationInput
	events        []DatabaseJobEventInput
	valid         bool
}

func CompleteDatabaseJob() DatabaseJobOutcome {
	outcome, _ := CompleteDatabaseJobWithEffects(DatabaseJobEffects{})
	return outcome
}

func ContinueDatabaseJob(
	successor DatabaseJobSuccessorInput,
) (DatabaseJobOutcome, error) {
	if !validDatabaseJobSuccessor(successor) {
		return DatabaseJobOutcome{}, ErrDatabaseJobOutcomeInvalid
	}
	return CompleteDatabaseJobWithEffects(DatabaseJobEffects{Successor: &successor})
}

func (outcome DatabaseJobOutcome) Successor() (DatabaseJobSuccessorInput, bool) {
	if !outcome.valid || !outcome.hasNext ||
		!validDatabaseJobSuccessor(outcome.successor) {
		return DatabaseJobSuccessorInput{}, false
	}
	return outcome.successor, true
}

func (outcome DatabaseJobOutcome) Valid() bool {
	if !outcome.valid {
		return false
	}
	var successor *DatabaseJobSuccessorInput
	if outcome.hasNext && !validDatabaseJobSuccessor(outcome.successor) {
		return false
	}
	if outcome.hasNext {
		successor = &outcome.successor
	}
	return validDatabaseJobCancellations(outcome.cancellations) &&
		validDatabaseJobChildren(outcome.children) &&
		!databaseJobChildConflicts(outcome.children, successor, outcome.cancellations) &&
		validDatabaseJobEvents(outcome.events)
}

func (outcome DatabaseJobOutcome) Children() []DatabaseJobChildInput {
	if !outcome.Valid() {
		return nil
	}
	return append([]DatabaseJobChildInput(nil), outcome.children...)
}

func (outcome DatabaseJobOutcome) Cancellations() []DatabaseJobCancellationInput {
	if !outcome.Valid() {
		return nil
	}
	return append([]DatabaseJobCancellationInput(nil), outcome.cancellations...)
}

func (outcome DatabaseJobOutcome) Events() []DatabaseJobEventInput {
	if !outcome.Valid() {
		return nil
	}
	result := make([]DatabaseJobEventInput, len(outcome.events))
	for index := range outcome.events {
		result[index] = outcome.events[index].clone()
	}
	return result
}

func validDatabaseJobSuccessor(value DatabaseJobSuccessorInput) bool {
	return value.JobID != uuid.Nil && value.JobID.Version() == uuid.Version(7) &&
		!value.RunAt.IsZero() && value.RunAt.Location() == time.UTC &&
		value.RunAt.Nanosecond()%int(time.Microsecond) == 0 &&
		validDatabaseJobText(value.PayloadReference, jobsv1.MaximumReferenceBytes) &&
		databaseJobDigest.MatchString(value.PayloadSHA256) &&
		validDatabaseJobText(value.DedupeKey, jobsv1.MaximumKeyBytes)
}

func validDatabaseJobText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value) &&
		strings.TrimSpace(value) == value
}
