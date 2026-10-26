package ports

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"strconv"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

type Mutation struct {
	IdempotencyKey  string
	ExpectedVersion int64
}

type MutationAction func(context.Context, Stores) (OperationReceipt, []byte, error)

type Mutator interface {
	Mutate(context.Context, Scope, string, domain.ID, Mutation, [32]byte,
		MutationAction) (OperationReceipt, error)
}

func Fingerprint(parts ...string) [32]byte {
	hash := sha256.New()
	var size [8]byte
	for _, part := range parts {
		binary.BigEndian.PutUint64(size[:], uint64(len(part)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write([]byte(part))
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func VersionString(version int64) string { return strconv.FormatInt(version, 10) }

func IDString(id *domain.ID) string {
	if id == nil {
		return ""
	}
	return id.String()
}
