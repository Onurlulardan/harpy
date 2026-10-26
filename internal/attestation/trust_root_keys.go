package attestation

import (
	"crypto/ed25519"
	"sort"
	"time"
)

// ActivePublicKeys returns defensive copies in stable key-ID order.
func (root TrustRoot) ActivePublicKeys(now time.Time) []ed25519.PublicKey {
	if now.IsZero() || now.Location() != time.UTC || len(root.keys) == 0 {
		return nil
	}
	ids := make([]string, 0, len(root.keys))
	for id, key := range root.keys {
		if !now.Before(key.notBefore) && now.Before(key.notAfter) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	keys := make([]ed25519.PublicKey, 0, len(ids))
	for _, id := range ids {
		keys = append(keys, append(ed25519.PublicKey(nil), root.keys[id].publicKey...))
	}
	return keys
}
