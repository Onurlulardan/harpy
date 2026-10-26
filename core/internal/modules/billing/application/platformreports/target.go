package platformreport

import "github.com/google/uuid"

type TransitionTarget struct {
	TenantID          uuid.UUID
	Provider          string
	GroupGeneration   int64
	BillingGeneration int64
	FormsGeneration   int64
}

func (target TransitionTarget) Valid() bool {
	return target.TenantID != uuid.Nil && target.Provider != "" &&
		target.GroupGeneration > 0 && target.BillingGeneration > 0 &&
		target.FormsGeneration > 0
}
