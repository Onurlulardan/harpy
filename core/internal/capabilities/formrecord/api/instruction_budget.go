package recordkernel

type instructionBudget struct {
	limit uint64
	used  uint64
}

func newInstructionBudget(catalog Catalog) instructionBudget {
	limit := catalog.limits.CalculationInstructions
	if limit > MaximumInstructions {
		limit = MaximumInstructions
	}
	return instructionBudget{limit: limit}
}

func (budget *instructionBudget) consume() error {
	if budget.used >= budget.limit {
		return ErrBudgetExceeded
	}
	budget.used++
	return nil
}
