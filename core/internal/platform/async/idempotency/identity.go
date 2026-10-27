package idempotency

import (
	"crypto/sha256"
	"encoding/binary"
	"regexp"

	"github.com/google/uuid"
)

var operationPattern = regexp.MustCompile(`^[a-z][a-z0-9_.:-]{0,127}$`)

type Identity struct {
	tenantID  uuid.UUID
	principal uuid.UUID
	operation string
	key       string
}

func NewIdentity(
	tenantID uuid.UUID,
	principalID uuid.UUID,
	operationID string,
	key string,
) (Identity, error) {
	if tenantID == uuid.Nil || principalID == uuid.Nil ||
		!operationPattern.MatchString(operationID) || len(key) < 1 || len([]byte(key)) > 512 {
		return Identity{}, ErrInvalidIdentity
	}
	return Identity{
		tenantID: tenantID, principal: principalID,
		operation: operationID, key: key,
	}, nil
}

func (identity Identity) TenantID() uuid.UUID    { return identity.tenantID }
func (identity Identity) PrincipalID() uuid.UUID { return identity.principal }
func (identity Identity) OperationID() string    { return identity.operation }
func (identity Identity) Key() string            { return identity.key }

func (identity Identity) AdvisoryKey() int64 {
	hash := sha256.New()
	for _, value := range []string{
		identity.tenantID.String(), identity.principal.String(),
		identity.operation, identity.key,
	} {
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(len([]byte(value))))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write([]byte(value))
	}
	return int64(binary.BigEndian.Uint64(hash.Sum(nil)[:8]))
}
