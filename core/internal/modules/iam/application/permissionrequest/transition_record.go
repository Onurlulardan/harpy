package permissionrequest

import "github.com/google/uuid"

type TransitionCommand string

const (
	TransitionCreate TransitionCommand = "iam.permission_request.create"
	TransitionReview TransitionCommand = "iam.permission_request.review"
	TransitionCancel TransitionCommand = "iam.permission_request.cancel"
	TransitionExpire TransitionCommand = "iam.permission_request.expire"
)

type TransitionRecord struct {
	OperationID   uuid.UUID
	RequestSHA256 [32]byte
	GroupID       uuid.UUID
	ActorID       uuid.UUID
	ReviewTaskID  uuid.UUID
	Command       TransitionCommand
	From          State
	To            State
	Revision      int64
}

func (record TransitionRecord) Validate() error {
	if record.OperationID == uuid.Nil || record.RequestSHA256 == [32]byte{} ||
		record.GroupID == uuid.Nil || record.ActorID == uuid.Nil || record.Revision < 1 ||
		!record.To.Valid() {
		return ErrInvalid
	}
	switch record.Command {
	case TransitionCreate:
		if record.From != "" || record.To != StatePending ||
			record.ReviewTaskID != uuid.Nil || record.Revision != 1 {
			return ErrInvalid
		}
	case TransitionReview:
		if record.From != StatePending || record.To != StateInReview ||
			record.ReviewTaskID == uuid.Nil {
			return ErrInvalid
		}
	case TransitionCancel:
		if !CanTransition(record.From, StateCancelled) ||
			record.To != StateCancelled || record.ReviewTaskID != uuid.Nil {
			return ErrInvalid
		}
	case TransitionExpire:
		if !CanTransition(record.From, StateExpired) ||
			record.To != StateExpired || record.ReviewTaskID != uuid.Nil {
			return ErrInvalid
		}
	default:
		return ErrInvalid
	}
	return nil
}
