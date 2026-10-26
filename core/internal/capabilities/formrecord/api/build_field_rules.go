package recordkernel

import "github.com/google/uuid"

func recordFieldPolicies(
	rules RecordRules,
	fieldCount int,
) (map[uuid.UUID]WritePolicy, error) {
	if len(rules.Fields) != fieldCount {
		return nil, ErrCatalogInvalid
	}
	result := make(map[uuid.UUID]WritePolicy, len(rules.Fields))
	for _, rule := range rules.Fields {
		fieldID, err := parseStableID(rule.FieldID)
		if err != nil || !rule.WritePolicy.valid() {
			return nil, ErrCatalogInvalid
		}
		if _, duplicate := result[fieldID]; duplicate {
			return nil, ErrCatalogInvalid
		}
		result[fieldID] = rule.WritePolicy
	}
	return result, nil
}
