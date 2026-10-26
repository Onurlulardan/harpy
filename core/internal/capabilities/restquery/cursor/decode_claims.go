package cursor

import (
	"bytes"
	"math"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/google/uuid"
)

func decodeClaims(plaintext []byte) (decodedClaims, error) {
	var wire wireClaims
	if attestation.DecodeStrict(plaintext, &wire) != nil {
		return decodedClaims{}, invalidCursorError()
	}
	canonical, err := attestation.Canonicalize(plaintext)
	defer clear(canonical)
	if err != nil || !bytes.Equal(canonical, plaintext) {
		return decodedClaims{}, invalidCursorError()
	}
	issuedAt, expiresAt, timeOK := decodeClaimTimes(wire.IssuedAt, wire.ExpiresAt)
	scopeKind, scopeID, subject, audience, operationID, snapshotAt, scopeOK :=
		decodeSnapshotIdentity(wire)
	authDigest, authOK := decodeDigest(wire.AuthorizationContextSHA256)
	queryDigest, queryOK := decodeDigest(wire.QueryIdentitySHA256)
	mandatoryDigest, mandatoryOK := decodeDigest(wire.MandatoryPredicateHMACSHA256)
	reference, referenceOK := decodeReference(wire)
	sorts, sortsOK := decodeWireSorts(wire.EffectiveSort)
	last, lastOK := decodeWireLastValues(sorts, wire.LastValues)
	if wire.Version != cursorVersion || !timeOK || !scopeOK ||
		!authOK || !queryOK || !mandatoryOK ||
		!referenceOK || !sortsOK || !lastOK || wire.PageSize < 1 ||
		wire.PageSize > maximumPageSize {
		return decodedClaims{}, invalidCursorError()
	}
	if scopeKind == "global" && !snapshotAt.Equal(issuedAt) {
		return decodedClaims{}, invalidCursorError()
	}
	return decodedClaims{
		issuedAt:             issuedAt,
		expiresAt:            expiresAt,
		scopeKind:            scopeKind,
		scopeID:              scopeID,
		subjectID:            subject,
		audience:             audience,
		operationID:          operationID,
		snapshotAt:           snapshotAt,
		authorizationContext: authDigest,
		reference:            reference,
		queryIdentity:        queryDigest,
		mandatoryDigest:      mandatoryDigest,
		pageSize:             wire.PageSize,
		sorts:                sorts,
		last:                 last,
	}, nil
}

func decodeSnapshotIdentity(
	wire wireClaims,
) (string, uuid.UUID, uuid.UUID, string, string, time.Time, bool) {
	if wire.TenantID != "" {
		scopeID, scopeOK := parseCanonicalUUID(wire.TenantID)
		subject, subjectOK := parseCanonicalUUID(wire.SubjectID)
		cleanGlobal := wire.ScopeKind == "" && wire.ScopeID == "" &&
			wire.Audience == "" && wire.OperationID == "" && wire.SnapshotAt == ""
		return "tenant", scopeID, subject, "tenant_actor", "restquery.execute",
			time.Time{}, scopeOK && subjectOK && cleanGlobal
	}
	if wire.ScopeKind != "global" || wire.ScopeID == "" || wire.TenantID != "" ||
		!globalOperationPattern.MatchString(wire.OperationID) {
		return "", uuid.Nil, uuid.Nil, "", "", time.Time{}, false
	}
	scopeID, scopeOK := parseCanonicalUUID(wire.ScopeID)
	snapshotAt, timeErr := time.Parse(time.RFC3339Nano, wire.SnapshotAt)
	if timeErr != nil || snapshotAt.Location() != time.UTC {
		return "", uuid.Nil, uuid.Nil, "", "", time.Time{}, false
	}
	if wire.Audience == "public" {
		return "global", scopeID, uuid.Nil, wire.Audience, wire.OperationID,
			snapshotAt, scopeOK && wire.SubjectID == ""
	}
	subject, subjectOK := parseCanonicalUUID(wire.SubjectID)
	return "global", scopeID, subject, wire.Audience, wire.OperationID,
		snapshotAt, scopeOK && subjectOK && wire.Audience == "authenticated"
}

func decodeClaimTimes(issuedAtUnix, expiresAtUnix int64) (time.Time, time.Time, bool) {
	if issuedAtUnix <= 0 || issuedAtUnix > math.MaxInt64-int64(cursorLifetime/time.Second) ||
		expiresAtUnix != issuedAtUnix+int64(cursorLifetime/time.Second) {
		return time.Time{}, time.Time{}, false
	}
	return time.Unix(issuedAtUnix, 0).UTC(), time.Unix(expiresAtUnix, 0).UTC(), true
}

func decodeReference(wire wireClaims) (api.CatalogReference, bool) {
	hash, err := api.NewCatalogHash(wire.CatalogSHA256)
	if err != nil {
		return api.CatalogReference{}, false
	}
	reference, err := api.NewCatalogReference(
		wire.ResourceOwner, wire.ResourceID, wire.PublishedVersionID, hash,
	)
	return reference, err == nil
}
