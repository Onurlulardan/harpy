package application

import "crypto/ed25519"

func testKey(marker byte) (ed25519.PublicKey, ed25519.PrivateKey) {
	seed := make([]byte, ed25519.SeedSize)
	for index := range seed {
		seed[index] = marker
	}
	private := ed25519.NewKeyFromSeed(seed)
	return private.Public().(ed25519.PublicKey), private
}

func testKeyPublic(marker byte) ed25519.PublicKey {
	public, _ := testKey(marker)
	return public
}

func keyArray(value ed25519.PublicKey) [32]byte {
	var result [32]byte
	copy(result[:], value)
	return result
}
