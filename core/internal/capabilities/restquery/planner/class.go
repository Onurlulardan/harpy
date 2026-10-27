package planner

import (
	"errors"
	"fmt"
)

var ErrStatementClassInvalid = errors.New("REST Query planner statement class is invalid")

// StatementClass identifies one planner-budgeted SQL statement role.
type StatementClass uint8

const (
	StatementClassRecord StatementClass = iota + 1
	StatementClassInclude
	StatementClassV1PredecessorSeek
	StatementClassExactCount
	StatementClassAggregate
)

func (class StatementClass) Valid() bool {
	return class >= StatementClassRecord && class <= StatementClassAggregate
}

func (class StatementClass) String() string {
	switch class {
	case StatementClassRecord:
		return "record"
	case StatementClassInclude:
		return "include"
	case StatementClassV1PredecessorSeek:
		return "v1_predecessor_seek"
	case StatementClassExactCount:
		return "exact_count"
	case StatementClassAggregate:
		return "aggregate"
	default:
		return "invalid"
	}
}

func (class StatementClass) GoString() string { return class.String() }

func (class StatementClass) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte(class.String()))
}
