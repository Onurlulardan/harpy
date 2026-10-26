package postgres

import (
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

type JobControl struct {
	owner     *workerIdentity
	authority JobAuthority
	jobID     uuid.UUID
}

func newJobControl(
	owner *workerIdentity,
	authority JobAuthority,
	jobID uuid.UUID,
) (JobControl, error) {
	control := JobControl{owner: owner, authority: authority, jobID: jobID}
	if !control.valid(owner) {
		return JobControl{}, ErrJobControlInvalid
	}
	return control, nil
}

func (control JobControl) valid(owner *workerIdentity) bool {
	return owner != nil && control.owner == owner &&
		control.authority.valid(owner) && control.jobID != uuid.Nil
}

type JobCancelResult struct {
	state       jobsv1.State
	requestedAt time.Time
}

func (result JobCancelResult) State() jobsv1.State    { return result.state }
func (result JobCancelResult) RequestedAt() time.Time { return result.requestedAt }
func (result JobCancelResult) Terminal() bool {
	return result.state == jobsv1.StateCancelled
}
