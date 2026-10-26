package cursor

import "time"

// SnapshotNow returns the canonical time used to authorize a first global page.
func (codec Codec) SnapshotNow() (time.Time, error) {
	if codec.Validate() != nil {
		return time.Time{}, invalidCursorError()
	}
	now := codec.state.clock().UTC().Truncate(time.Second)
	if now.IsZero() || now.Unix() <= 0 {
		return time.Time{}, invalidCursorError()
	}
	return now, nil
}

// AuthenticatedGlobalSnapshotAt recovers only the authenticated global snapshot
// anchor. Full query, authority, audience and operation binding still happens in Decode.
func (codec Codec) AuthenticatedGlobalSnapshotAt(token string) (time.Time, error) {
	if codec.Validate() != nil {
		return time.Time{}, invalidCursorError()
	}
	plaintext, _, err := codec.decryptToken(token)
	if err != nil {
		return time.Time{}, err
	}
	defer clear(plaintext)
	decoded, err := decodeClaims(plaintext)
	if err != nil || decoded.scopeKind != "global" {
		return time.Time{}, invalidCursorError()
	}
	if err := validateDecodedTime(codec.state.clock().UTC(), decoded); err != nil {
		return time.Time{}, err
	}
	return decoded.snapshotAt, nil
}

func validateDecodedTime(now time.Time, decoded decodedClaims) error {
	if now.IsZero() || now.Unix() <= 0 ||
		decoded.issuedAt.After(now.Add(allowedClockSkew)) {
		return invalidCursorError()
	}
	if now.After(decoded.expiresAt.Add(allowedClockSkew)) {
		return expiredCursorError()
	}
	return nil
}
