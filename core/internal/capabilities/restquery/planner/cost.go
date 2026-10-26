package planner

import (
	"errors"
	"math"
	"strconv"
)

var ErrCostInvalid = errors.New("REST Query planner cost is invalid")

// Cost is an immutable, finite PostgreSQL planner cost.
type Cost struct {
	value float64
}

func NewCost(value float64) (Cost, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || math.Signbit(value) {
		return Cost{}, ErrCostInvalid
	}
	return Cost{value: value}, nil
}

func (cost Cost) Value() float64 { return cost.value }

func (cost Cost) String() string {
	if !cost.valid() {
		return "invalid"
	}
	return strconv.FormatFloat(cost.value, 'g', -1, 64)
}

func (cost Cost) valid() bool {
	return !math.IsNaN(cost.value) && !math.IsInf(cost.value, 0) && !math.Signbit(cost.value)
}
