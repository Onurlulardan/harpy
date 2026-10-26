package config

import (
	"crypto/ed25519"
	"strings"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

func newStringReader(value string) *strings.Reader { return strings.NewReader(value) }

func actorBridgeVerifier(
	keys map[string]ed25519.PublicKey,
) (actorassertionv1.Verifier, error) {
	return actorassertionv1.NewVerifier(keys)
}

func iamActorBridgeKeys() []string {
	return []string{
		iamActorBridgeKeyIDKey,
		iamActorBridgeSeedFileKey,
		iamActorBridgePublicKeysFileKey,
		iamActorBridgeTargetsFileKey,
		iamActorBridgeClientCertKey,
		iamActorBridgeClientKeyKey,
		iamActorBridgeServerCAKey,
	}
}
