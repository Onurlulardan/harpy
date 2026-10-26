package async

import (
	"regexp"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

var databaseJobKind = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,127}$`)

// DatabaseJobChildInput is a different-kind job requested by a database
// command. Tenant, authority, effect class, attempts and retention policy are
// inherited or resolved by the platform and are intentionally absent here.
type DatabaseJobChildInput struct {
	JobID                uuid.UUID
	Kind                 string
	PayloadRevision      int
	RunAt                time.Time
	PayloadReference     string
	PayloadSHA256        string
	DedupeKey            string
	EffectIdempotencyKey string
	Priority             int32
}

func (value DatabaseJobChildInput) Valid() bool {
	return validDatabaseJobChild(value)
}

func validDatabaseJobChild(value DatabaseJobChildInput) bool {
	return value.JobID != uuid.Nil && value.JobID.Version() == uuid.Version(7) &&
		databaseJobKind.MatchString(value.Kind) && value.PayloadRevision > 0 &&
		!value.RunAt.IsZero() && value.RunAt.Location() == time.UTC &&
		value.RunAt.Nanosecond()%int(time.Microsecond) == 0 &&
		validDatabaseJobText(value.PayloadReference, jobsv1.MaximumReferenceBytes) &&
		databaseJobDigest.MatchString(value.PayloadSHA256) &&
		validDatabaseJobText(value.DedupeKey, jobsv1.MaximumKeyBytes) &&
		(value.EffectIdempotencyKey == "" || validDatabaseJobText(
			value.EffectIdempotencyKey, jobsv1.MaximumKeyBytes,
		))
}

func validDatabaseJobChildren(values []DatabaseJobChildInput) bool {
	identities := map[uuid.UUID]bool{}
	dedupe := map[string]bool{}
	for _, value := range values {
		key := value.Kind + "\x00" + value.DedupeKey
		if !validDatabaseJobChild(value) || identities[value.JobID] || dedupe[key] {
			return false
		}
		identities[value.JobID], dedupe[key] = true, true
	}
	return true
}

func databaseJobChildConflicts(
	children []DatabaseJobChildInput,
	successor *DatabaseJobSuccessorInput,
	cancellations []DatabaseJobCancellationInput,
) bool {
	for _, child := range children {
		if successor != nil && child.JobID == successor.JobID ||
			cancelsDatabaseJob(cancellations, child.JobID) {
			return true
		}
	}
	return false
}
