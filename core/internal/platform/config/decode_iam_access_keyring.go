package config

import (
	"crypto/ed25519"
	"encoding/json"
)

func decodeIAMRetiredAccessKey(raw json.RawMessage) (iamRetiredAccessKey, bool) {
	object, ok := decodeIAMObject(raw, "kid", "public_key", "sign_until", "publish_until")
	if !ok {
		return iamRetiredAccessKey{}, false
	}
	kid, publicText, signText, publishText, ok := decodeIAMStrings(
		object, "kid", "public_key", "sign_until", "publish_until",
	)
	publicMaterial, publicOK := decodeIAMKeyMaterial(publicText)
	signUntil, signOK := parseIAMTimestamp(signText)
	publishUntil, publishOK := parseIAMTimestamp(publishText)
	if !ok || !validIAMAccessKeyID(kid) || !publicOK || iamAllZero(publicMaterial[:]) ||
		!signOK || !publishOK || !signUntil.Before(publishUntil) {
		return iamRetiredAccessKey{}, false
	}
	publicKey := append(ed25519.PublicKey(nil), publicMaterial[:]...)
	clear(publicMaterial[:])
	return iamRetiredAccessKey{
		kid: kid, publicKey: publicKey,
		signUntil: signUntil, publishUntil: publishUntil,
	}, true
}

func iamAccessPublicMatchesSeed(publicKey ed25519.PublicKey, seed [32]byte) bool {
	privateKey := ed25519.NewKeyFromSeed(seed[:])
	current := privateKey.Public().(ed25519.PublicKey)
	matches := string(publicKey) == string(current)
	clear(privateKey)
	return matches
}
