package iam

import (
	"crypto/sha256"
	"encoding/binary"
	"time"

	"github.com/google/uuid"
)

func catalogAuthorizationEvidenceSHA256(
	principalID, tenantID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permissionID uuid.UUID,
	permission, operation string,
	target catalogAccessTarget,
	stepUpAt time.Time,
	stepUpPurpose [sha256.Size]byte,
) [sha256.Size]byte {
	hash := sha256.New()
	writeCatalogEvidenceField(hash, []byte("iam.catalog-authorization.v1"))
	writeCatalogEvidenceField(hash, principalID[:])
	writeCatalogEvidenceField(hash, tenantID[:])
	writeCatalogEvidenceInteger(hash, tokenVersion)
	writeCatalogEvidenceInteger(hash, authorizationVersion)
	writeCatalogEvidenceInteger(hash, authorizationClock)
	writeCatalogEvidenceField(hash, permissionID[:])
	writeCatalogEvidenceField(hash, []byte(permission))
	writeCatalogEvidenceField(hash, []byte(operation))
	writeCatalogEvidenceField(hash, []byte(target.resource))
	writeCatalogEvidenceField(hash, []byte(target.action))
	writeCatalogEvidenceField(hash, []byte(stepUpAt.UTC().Format(time.RFC3339Nano)))
	writeCatalogEvidenceField(hash, stepUpPurpose[:])
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func writeCatalogEvidenceInteger(hash interface{ Write([]byte) (int, error) }, value int64) {
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], uint64(value))
	writeCatalogEvidenceField(hash, encoded[:])
}

func writeCatalogEvidenceField(hash interface{ Write([]byte) (int, error) }, value []byte) {
	var size [4]byte
	binary.BigEndian.PutUint32(size[:], uint32(len(value)))
	_, _ = hash.Write(size[:])
	_, _ = hash.Write(value)
}
