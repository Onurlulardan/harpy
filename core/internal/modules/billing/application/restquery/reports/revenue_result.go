package reports

import "encoding/json"

type RevenueResult struct {
	groups []json.RawMessage
	limit  int
}

func newRevenueResult(groups [][]byte, limit int) RevenueResult {
	result := RevenueResult{groups: make([]json.RawMessage, len(groups)), limit: limit}
	for index, group := range groups {
		result.groups[index] = append(json.RawMessage(nil), group...)
	}
	return result
}

func (result RevenueResult) Groups() []json.RawMessage {
	groups := make([]json.RawMessage, len(result.groups))
	for index, group := range result.groups {
		groups[index] = append(json.RawMessage(nil), group...)
	}
	return groups
}

func (result RevenueResult) Limit() int { return result.limit }
