package application

import "crypto/sha256"

func hashContent(value string) [32]byte {
	return sha256.Sum256([]byte(value))
}
