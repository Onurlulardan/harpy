package catalogprojection

func NewBoundStore(
	owner string,
	contract string,
	generation int64,
	query string,
	kinds []string,
	queryRow QueryRow,
	allowed ScopeGuard,
) (*BoundStore, error) {
	kindSet := make(map[string]bool, len(kinds))
	for _, kind := range kinds {
		if kind == "" || len(kind) > 32 || kindSet[kind] {
			return nil, ErrInvalid
		}
		kindSet[kind] = true
	}
	if !ownerPattern.MatchString(owner) || contract == "" || len(contract) > 64 ||
		generation < 1 || query == "" || len(kindSet) == 0 ||
		queryRow == nil || allowed == nil {
		return nil, ErrInvalid
	}
	return &BoundStore{
		owner: owner, contract: contract, generation: generation,
		query: query, kinds: kindSet, queryRow: queryRow, allowed: allowed,
	}, nil
}
