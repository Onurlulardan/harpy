package sqlidentifier

import "errors"

// MintStatic validates an owner-catalogued static identifier.
func MintStatic(value string) (Token, error) {
	if !validLowerIdentifier(value) || hasDynamicPrefix(value) {
		return Token{}, errors.New("static SQL identifier is invalid")
	}
	return Token{identifier: value, kind: tokenKindStatic}, nil
}
