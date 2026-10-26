package discount

import "errors"

var (
	ErrCode     = errors.New("billing discount code is invalid")
	ErrDuration = errors.New("billing discount duration is invalid")
	ErrRule     = errors.New("billing discount rule is invalid")
	ErrApply    = errors.New("billing discount application is invalid")
	ErrPolicy   = errors.New("billing discount policy is invalid")
)
