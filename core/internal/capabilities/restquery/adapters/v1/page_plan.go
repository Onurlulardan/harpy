package v1query

import (
	"fmt"
	"math"
)

type PagePlan struct {
	page             int64
	limit            int
	predecessorIndex int64
	hasPredecessor   bool
	validated        bool
}

func newPagePlan(page int64, limit int) (PagePlan, error) {
	if page < 1 || limit < 1 || limit > 500 {
		return PagePlan{}, fmt.Errorf("V1 page plan bounds are invalid")
	}
	plan := PagePlan{page: page, limit: limit, validated: true}
	if page > 1 {
		if page-1 > math.MaxInt64/int64(limit) {
			return PagePlan{}, fmt.Errorf("V1 page ordinal overflows")
		}
		plan.predecessorIndex = (page-1)*int64(limit) - 1
		plan.hasPredecessor = true
	}
	if err := plan.Validate(); err != nil {
		return PagePlan{}, err
	}
	return plan, nil
}

func (plan PagePlan) Page() (int64, error) {
	if err := plan.Validate(); err != nil {
		return 0, err
	}
	return plan.page, nil
}

func (plan PagePlan) Limit() (int, error) {
	if err := plan.Validate(); err != nil {
		return 0, err
	}
	return plan.limit, nil
}

func (plan PagePlan) PredecessorIndex() (int64, bool, error) {
	if err := plan.Validate(); err != nil {
		return 0, false, err
	}
	return plan.predecessorIndex, plan.hasPredecessor, nil
}

func (plan PagePlan) Validate() error {
	if !plan.validated || plan.page < 1 || plan.limit < 1 || plan.limit > 500 ||
		plan.hasPredecessor != (plan.page > 1) ||
		!plan.hasPredecessor && plan.predecessorIndex != 0 {
		return fmt.Errorf("V1 REST Query page plan is invalid")
	}
	if plan.hasPredecessor {
		if plan.page-1 > math.MaxInt64/int64(plan.limit) ||
			plan.predecessorIndex != (plan.page-1)*int64(plan.limit)-1 {
			return fmt.Errorf("V1 REST Query page plan is invalid")
		}
	}
	return nil
}

func (PagePlan) String() string                      { return v1Redaction }
func (PagePlan) GoString() string                    { return v1Redaction }
func (plan PagePlan) Format(state fmt.State, _ rune) { writeV1Redaction(state) }
func (PagePlan) MarshalJSON() ([]byte, error)        { return rejectV1JSON() }
