package sqlidentifier

import "errors"

// MintDynamic validates an identifier in a Forms-owned dynamic namespace.
func MintDynamic(value string) (Token, error) {
	if !validLowerIdentifier(value) || !hasDynamicPrefix(value) {
		return Token{}, errors.New("dynamic SQL identifier is invalid")
	}
	return Token{identifier: value, kind: tokenKindDynamic}, nil
}
