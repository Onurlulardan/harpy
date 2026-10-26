package subscription

import "errors"

var (
	ErrState      = errors.New("billing subscription state is invalid")
	ErrTransition = errors.New("billing subscription transition is invalid")
	ErrGrant      = errors.New("billing subscription grant window is invalid")
	ErrSchedule   = errors.New("billing subscription schedule is invalid")
	ErrChange     = errors.New("billing subscription manual change is invalid")
)
