package cursor

import (
	"crypto/sha256"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

type decodedClaims struct {
	issuedAt             time.Time
	expiresAt            time.Time
	scopeKind            string
	scopeID              uuid.UUID
	subjectID            uuid.UUID
	audience             string
	operationID          string
	snapshotAt           time.Time
	authorizationContext [sha256.Size]byte
	reference            api.CatalogReference
	queryIdentity        [sha256.Size]byte
	mandatoryDigest      [sha256.Size]byte
	pageSize             int
	sorts                []decodedSort
	last                 []LastValue
}

type decodedSort struct {
	fieldID     uuid.UUID
	direction   api.Direction
	logicalType api.LogicalType
}
