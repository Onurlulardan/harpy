package cursor

// Decode authenticates cursor binding only. The caller must authorize every page again.
func (codec Codec) Decode(token string, expected Context) (Claims, error) {
	if codec.Validate() != nil || expected.Validate() != nil {
		return Claims{}, invalidCursorError()
	}
	plaintext, key, err := codec.decryptToken(token)
	if err != nil {
		return Claims{}, err
	}
	defer clear(plaintext)
	decoded, err := decodeClaims(plaintext)
	if err != nil {
		return Claims{}, err
	}
	if err := validateDecodedTime(codec.state.clock().UTC(), decoded); err != nil {
		return Claims{}, err
	}
	switch bindExpected(decoded, expected, key) {
	case bindingInvalid:
		return Claims{}, invalidCursorError()
	case bindingStale:
		return Claims{}, staleCursorError()
	}
	claims, claimsErr := newClaims(decoded.issuedAt, decoded.expiresAt, decoded.last)
	if claimsErr != nil {
		return Claims{}, invalidCursorError()
	}
	return claims, nil
}
