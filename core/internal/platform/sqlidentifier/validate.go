package sqlidentifier

import "errors"

var errInvalidToken = errors.New("SQL identifier token is invalid")

// Validate proves that the opaque token still satisfies its minting rules.
func (token Token) Validate() error {
	if !validLowerIdentifier(token.identifier) {
		return errInvalidToken
	}
	switch token.kind {
	case tokenKindStatic:
		if hasDynamicPrefix(token.identifier) {
			return errInvalidToken
		}
	case tokenKindDynamic:
		if !hasDynamicPrefix(token.identifier) {
			return errInvalidToken
		}
	default:
		return errInvalidToken
	}
	return nil
}
