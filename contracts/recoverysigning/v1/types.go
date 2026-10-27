package recoverysigningv1

import "time"

type PublicMaterial struct {
	Purpose   Purpose
	PublicKey [32]byte
}

type keyWire struct {
	Purpose     Purpose `json:"purpose"`
	KeyID       string  `json:"key_id"`
	Algorithm   string  `json:"algorithm"`
	Status      string  `json:"status"`
	PublicKey   string  `json:"public_key"`
	NotBefore   string  `json:"not_before"`
	SignUntil   *string `json:"sign_until"`
	VerifyUntil *string `json:"verify_until"`
}

type bundleWire struct {
	Schema        string    `json:"schema"`
	EnvironmentID string    `json:"environment_id"`
	Generation    int64     `json:"generation"`
	Algorithm     string    `json:"algorithm"`
	GeneratedAt   string    `json:"generated_at"`
	Keys          []keyWire `json:"keys"`
	Retired       []keyWire `json:"retired,omitempty"`
}

type Bundle struct {
	wire      bundleWire
	keys      map[Purpose]keyMaterial
	retired   map[string]keyMaterial
	canonical []byte
	digest    string
	createdAt time.Time
}

type keyMaterial struct {
	id        string
	purpose   Purpose
	publicKey [32]byte
}

func (bundle Bundle) EnvironmentID() string  { return bundle.wire.EnvironmentID }
func (bundle Bundle) Generation() int64      { return bundle.wire.Generation }
func (bundle Bundle) GeneratedAt() time.Time { return bundle.createdAt }
func (bundle Bundle) SHA256() string         { return bundle.digest }
func (bundle Bundle) CanonicalBytes() []byte {
	return append([]byte(nil), bundle.canonical...)
}
func (bundle Bundle) PublicKey(purpose Purpose) ([32]byte, bool) {
	value, found := bundle.keys[purpose]
	return value.publicKey, found
}
func (bundle Bundle) KeyID(purpose Purpose) (string, bool) {
	value, found := bundle.keys[purpose]
	return value.id, found
}
func (bundle Bundle) Retains(purpose Purpose, keyID, publicKeySHA256 string) bool {
	value, found := bundle.retired[keyID]
	return found && value.purpose == purpose && digestPublicKey(value.publicKey) == publicKeySHA256
}

func (bundle Bundle) VerificationKey(
	purpose Purpose,
	keyID string,
) ([32]byte, int64, bool) {
	if active, found := bundle.keys[purpose]; found && active.id == keyID {
		return active.publicKey, bundle.Generation(), true
	}
	retired, found := bundle.retired[keyID]
	if !found || retired.purpose != purpose {
		return [32]byte{}, 0, false
	}
	generation, valid := keyGeneration(keyID, purpose, retired.publicKey)
	return retired.publicKey, generation, valid
}
