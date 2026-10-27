package accesstoken

import "strings"

type compactSegments struct {
	header, payload, signature string
}

func splitCompact(token string) (compactSegments, bool) {
	if token == "" || len(token) > MaximumWireSize || strings.Count(token, ".") != 2 {
		return compactSegments{}, false
	}
	first := strings.IndexByte(token, '.')
	secondRelative := strings.IndexByte(token[first+1:], '.')
	if first <= 0 || secondRelative <= 0 {
		return compactSegments{}, false
	}
	second := first + 1 + secondRelative
	if second == len(token)-1 {
		return compactSegments{}, false
	}
	return compactSegments{
		header: token[:first], payload: token[first+1 : second], signature: token[second+1:],
	}, true
}
