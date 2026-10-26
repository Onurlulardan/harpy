package cache

import (
	"crypto/sha256"
	"encoding/binary"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

func deriveKey(request queryapp.CacheRequest) ([32]byte, bool) {
	if request.TenantID == uuid.Nil || request.QueryID == uuid.Nil ||
		request.QueryVersionID == uuid.Nil || request.Binding.ReleaseGeneration == 0 ||
		request.Binding.AuthorizationSHA256 == [32]byte{} ||
		request.Binding.PredicateSHA256 == [32]byte{} ||
		request.Binding.RevisionsSHA256 == [32]byte{} ||
		request.QuerySHA256 == [32]byte{} || request.ParametersSHA256 == [32]byte{} {
		return [32]byte{}, false
	}
	digest := sha256.New()
	_, _ = digest.Write(request.TenantID[:])
	_, _ = digest.Write(request.QueryID[:])
	_, _ = digest.Write(request.QueryVersionID[:])
	generation := make([]byte, 8)
	binary.BigEndian.PutUint64(generation, request.Binding.ReleaseGeneration)
	_, _ = digest.Write(generation)
	for _, value := range [][32]byte{request.Binding.AuthorizationSHA256,
		request.Binding.PredicateSHA256, request.Binding.RevisionsSHA256,
		request.QuerySHA256, request.ParametersSHA256} {
		_, _ = digest.Write(value[:])
	}
	var result [32]byte
	copy(result[:], digest.Sum(nil))
	return result, true
}
