package discount

import "bytes"

func validApplicability(all bool, identifiers []Identifier) bool {
	if all {
		return len(identifiers) == 0
	}
	if len(identifiers) == 0 || len(identifiers) > 256 {
		return false
	}
	seen := make(map[Identifier]struct{}, len(identifiers))
	for _, identifier := range identifiers {
		if zeroIdentifier(identifier) {
			return false
		}
		if _, exists := seen[identifier]; exists {
			return false
		}
		seen[identifier] = struct{}{}
	}
	return true
}

func compareIdentifier(left, right Identifier) int {
	return bytes.Compare(left[:], right[:])
}
