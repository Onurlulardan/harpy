package revisioncache

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
)

type Binding struct {
	TenantID            [16]byte
	ReleaseID           [16]byte
	ReleaseGeneration   uint64
	QuerySHA256         [32]byte
	ParametersSHA256    [32]byte
	AuthorizationSHA256 [32]byte
	PredicateSHA256     [32]byte
	RevisionsSHA256     [32]byte
}

func (binding Binding) Validate() error {
	if zero16(binding.TenantID) || zero16(binding.ReleaseID) ||
		binding.ReleaseGeneration == 0 || zero32(binding.QuerySHA256) ||
		zero32(binding.ParametersSHA256) || zero32(binding.AuthorizationSHA256) ||
		zero32(binding.PredicateSHA256) || zero32(binding.RevisionsSHA256) {
		return ErrInvalid
	}
	return nil
}

func (binding Binding) Key() ([32]byte, error) {
	if err := binding.Validate(); err != nil {
		return [32]byte{}, err
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte("harpy/analytics/revision-cache/v1"))
	_, _ = hash.Write(binding.TenantID[:])
	_, _ = hash.Write(binding.ReleaseID[:])
	generation := make([]byte, 8)
	binary.BigEndian.PutUint64(generation, binding.ReleaseGeneration)
	_, _ = hash.Write(generation)
	for _, value := range [][32]byte{binding.QuerySHA256, binding.ParametersSHA256,
		binding.AuthorizationSHA256, binding.PredicateSHA256, binding.RevisionsSHA256} {
		_, _ = hash.Write(value[:])
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}

func zero16(value [16]byte) bool { return bytes.Equal(value[:], make([]byte, 16)) }
func zero32(value [32]byte) bool { return bytes.Equal(value[:], make([]byte, 32)) }
