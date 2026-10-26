package licenseauthority

import "time"

func parseKeysetPayload(
	document []byte, now time.Time,
) (parsedTrustedIssuer, time.Time, error) {
	var wire keysetPayloadWire
	if strictDecode(document, &wire) != nil || wire.Schema != KeysetSchema ||
		wire.Audience != KeysetAudience || !validText(wire.Issuer, 256) ||
		wire.Sequence < 1 || len(wire.SnapshotSigners) == 0 || len(wire.SnapshotSigners) > 64 {
		return parsedTrustedIssuer{}, time.Time{}, ErrInvalidTrustBundle
	}
	issuedAt, valid := parseExactInstant(wire.IssuedAt)
	if !valid {
		return parsedTrustedIssuer{}, time.Time{}, ErrInvalidTrustBundle
	}
	notBefore, valid := parseExactInstant(wire.NotBefore)
	if !valid {
		return parsedTrustedIssuer{}, time.Time{}, ErrInvalidTrustBundle
	}
	expiresAt, valid := parseExactInstant(wire.ExpiresAt)
	if !valid || issuedAt.After(notBefore) || !notBefore.Before(expiresAt) ||
		now.Before(notBefore) || now.After(expiresAt) {
		return parsedTrustedIssuer{}, time.Time{}, ErrInvalidTrustBundle
	}
	keys := make(map[string]trustedLicenseKey, len(wire.SnapshotSigners))
	previous := ""
	for _, value := range wire.SnapshotSigners {
		if previous != "" && value.KeyID <= previous {
			return parsedTrustedIssuer{}, time.Time{}, ErrInvalidTrustBundle
		}
		key, err := parseTrustKey(value.KeyID, value.Algorithm, value.PublicKey,
			value.Status, value.NotBefore, value.SignUntil, value.VerifyUntil)
		if err != nil {
			return parsedTrustedIssuer{}, time.Time{}, err
		}
		keys[value.KeyID], previous = key, value.KeyID
	}
	return parsedTrustedIssuer{
		name:          wire.Issuer,
		trustedIssuer: trustedIssuer{sequence: wire.Sequence, keys: keys},
	}, issuedAt, nil
}
