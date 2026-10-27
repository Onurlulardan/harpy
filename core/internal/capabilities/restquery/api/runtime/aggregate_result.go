package runtime

type AggregateResult struct {
	groups [][]byte
	limit  int
}

func newAggregateResult(groups [][]byte, limit int) AggregateResult {
	return AggregateResult{groups: cloneItems(groups), limit: limit}
}

func (result AggregateResult) Groups() [][]byte { return cloneItems(result.groups) }
func (result AggregateResult) Limit() int       { return result.limit }
