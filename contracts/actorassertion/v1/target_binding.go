package actorassertionv1

import (
	"crypto/sha256"
	"encoding/binary"
)

const targetBindingDomain = "harpy:iam-v1-actor-target:v1"

func TargetBindingSHA256(parts ...string) ([32]byte, error) {
	if len(parts) < 1 || len(parts) > 16 {
		return [32]byte{}, ErrContract
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(targetBindingDomain))
	for _, part := range parts {
		if len(part) < 1 || len(part) > 1024 {
			return [32]byte{}, ErrContract
		}
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], uint32(len(part)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write([]byte(part))
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}
