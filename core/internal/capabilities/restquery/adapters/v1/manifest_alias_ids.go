package v1query

import "slices"

func (policy manifest) aliasIDs() []string {
	result := make([]string, 0)
	for _, binding := range policy.filterFields {
		result = appendAliasID(result, binding.aliasID)
	}
	for _, binding := range policy.sortFields {
		result = appendAliasID(result, binding.aliasID)
	}
	for _, binding := range policy.operatorAliases {
		result = appendAliasID(result, binding.aliasID)
	}
	for _, binding := range policy.queryKeyAliases {
		result = appendAliasID(result, binding.aliasID)
	}
	slices.Sort(result)
	return result
}

func appendAliasID(values []string, aliasID string) []string {
	if aliasID == "" {
		return values
	}
	return append(values, aliasID)
}
