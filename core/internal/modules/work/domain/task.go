package domain

import "errors"

var (
	ErrInvalidTransition = errors.New("work task transition is invalid")
	ErrVersionConflict   = errors.New("work row version conflicts")
)

type TaskKind string

const (
	TaskCustom     TaskKind = "CUSTOM"
	TaskFormFill   TaskKind = "FORM_FILL"
	TaskFormReview TaskKind = "FORM_REVIEW"
)

func (kind TaskKind) Valid() bool {
	return kind == TaskCustom || kind == TaskFormFill || kind == TaskFormReview
}

type TaskStatus string

const (
	StatusOpen       TaskStatus = "OPEN"
	StatusInProgress TaskStatus = "IN_PROGRESS"
	StatusInReview   TaskStatus = "IN_REVIEW"
	StatusDone       TaskStatus = "DONE"
	StatusCancelled  TaskStatus = "CANCELLED"
)

func (status TaskStatus) Valid() bool {
	switch status {
	case StatusOpen, StatusInProgress, StatusInReview, StatusDone, StatusCancelled:
		return true
	default:
		return false
	}
}

func (status TaskStatus) Terminal() bool {
	return status == StatusDone || status == StatusCancelled
}

type TaskPriority string

const (
	PriorityLow    TaskPriority = "LOW"
	PriorityMedium TaskPriority = "MEDIUM"
	PriorityHigh   TaskPriority = "HIGH"
	PriorityUrgent TaskPriority = "URGENT"
)

func (priority TaskPriority) Valid() bool {
	switch priority {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

type Transition string

const (
	TransitionMove     Transition = "MOVE"
	TransitionComplete Transition = "COMPLETE"
	TransitionCancel   Transition = "CANCEL"
	TransitionReopen   Transition = "REOPEN"
)

func NextStatus(current, target TaskStatus, transition Transition) (TaskStatus, error) {
	if !current.Valid() || !target.Valid() {
		return "", ErrInvalidTransition
	}
	switch transition {
	case TransitionMove:
		if current.Terminal() || target == StatusCancelled {
			return "", ErrInvalidTransition
		}
		return target, nil
	case TransitionComplete:
		if current.Terminal() || target != StatusDone {
			return "", ErrInvalidTransition
		}
		return StatusDone, nil
	case TransitionCancel:
		if current.Terminal() || target != StatusCancelled {
			return "", ErrInvalidTransition
		}
		return StatusCancelled, nil
	case TransitionReopen:
		if !current.Terminal() || target != StatusOpen {
			return "", ErrInvalidTransition
		}
		return StatusOpen, nil
	default:
		return "", ErrInvalidTransition
	}
}
