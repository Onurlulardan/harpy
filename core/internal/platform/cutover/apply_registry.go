package cutover

import "fmt"

type ApplyRegistry struct {
	queries map[string]string
}

func NewApplyRegistry(contracts ...ApplyContract) (ApplyRegistry, error) {
	if len(contracts) == 0 {
		return ApplyRegistry{}, fmt.Errorf("cutover apply registry is empty")
	}
	queries := make(map[string]string, len(contracts))
	for _, contract := range contracts {
		if contract.ID() == "" || contract.Query() == "" {
			return ApplyRegistry{}, fmt.Errorf("cutover apply contract is invalid")
		}
		if _, duplicate := queries[contract.ID()]; duplicate {
			return ApplyRegistry{}, fmt.Errorf("cutover apply contract is duplicated")
		}
		queries[contract.ID()] = contract.Query()
	}
	return ApplyRegistry{queries: queries}, nil
}

func (registry ApplyRegistry) Lookup(contractID string) (string, bool) {
	query, found := registry.queries[contractID]
	return query, found
}

func (registry ApplyRegistry) Valid() bool {
	return len(registry.queries) > 0
}
