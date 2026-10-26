package postgres

import (
	"time"

	"github.com/google/uuid"
)

type OwnedJobCancellation struct {
	jobID        uuid.UUID
	requestedAt  time.Time
	cancellation JobCancellation
}

func NewOwnedJobCancellation(
	jobID uuid.UUID,
	requestedAt time.Time,
	code string,
	digest string,
) (OwnedJobCancellation, error) {
	cancellation, err := NewJobCancellation(code, digest)
	if err != nil || jobID == uuid.Nil || jobID.Version() != uuid.Version(7) ||
		requestedAt.IsZero() || requestedAt.Location() != time.UTC ||
		requestedAt.Nanosecond()%int(time.Microsecond) != 0 {
		return OwnedJobCancellation{}, ErrJobFailureInvalid
	}
	return OwnedJobCancellation{
		jobID: jobID, requestedAt: requestedAt, cancellation: cancellation,
	}, nil
}

func (value OwnedJobCancellation) valid() bool {
	return value.jobID != uuid.Nil && value.jobID.Version() == uuid.Version(7) &&
		!value.requestedAt.IsZero() && value.requestedAt.Location() == time.UTC &&
		value.requestedAt.Nanosecond()%int(time.Microsecond) == 0 &&
		value.cancellation.valid()
}
