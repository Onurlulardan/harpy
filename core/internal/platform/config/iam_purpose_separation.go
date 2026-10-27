package config

import "bytes"

func iamPurposeKeysUnique(
	access IAMAccessTokenKeyring,
	rings []iamHMACKeyring,
	actorBridge IAMActorBridge,
	metadataIDs ...string,
) bool {
	seenIDs := make(map[string]struct{})
	seenMaterials := make([][32]byte, 0)
	if !addIAMPurposeMaterials(access.keyMaterials(), seenIDs, &seenMaterials) {
		return false
	}
	for _, ring := range rings {
		if !addIAMPurposeMaterials(ring.keyMaterials(), seenIDs, &seenMaterials) {
			return false
		}
	}
	if len(actorBridge.privateKey) != 64 {
		return false
	}
	var actorSeed [32]byte
	copy(actorSeed[:], actorBridge.privateKey[:32])
	if !addIAMPurposeMaterials(
		map[string][32]byte{actorBridge.keyID: actorSeed}, seenIDs, &seenMaterials,
	) {
		return false
	}
	for _, identifier := range metadataIDs {
		if _, duplicate := seenIDs[identifier]; duplicate {
			return false
		}
		seenIDs[identifier] = struct{}{}
	}
	return true
}

func addIAMPurposeMaterials(
	materials map[string][32]byte,
	seenIDs map[string]struct{},
	seenMaterials *[][32]byte,
) bool {
	for kid, material := range materials {
		if _, duplicate := seenIDs[kid]; duplicate {
			return false
		}
		for _, existing := range *seenMaterials {
			if bytes.Equal(existing[:], material[:]) {
				return false
			}
		}
		seenIDs[kid] = struct{}{}
		*seenMaterials = append(*seenMaterials, material)
	}
	return true
}
