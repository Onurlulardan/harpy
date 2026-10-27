package cursor

import (
	"time"

	"github.com/google/uuid"
)

func (codec Codec) buildWireClaims(
	context Context,
	lastValues []LastValue,
	issuedAt time.Time,
) (wireClaims, error) {
	owner, ownerErr := context.catalog.Owner()
	resourceID, resourceErr := context.catalog.ResourceID()
	publishedVersionID, versionErr := context.catalog.PublishedVersionID()
	catalogHash, hashErr := context.catalog.CatalogHash()
	catalogHex, catalogHexErr := catalogHash.Hex()
	last, lastErr := encodeWireLastValues(lastValues)
	if ownerErr != nil || resourceErr != nil || versionErr != nil || hashErr != nil ||
		catalogHexErr != nil || lastErr != nil {
		return wireClaims{}, errEncodeFailed
	}
	mandatoryDigest := mandatoryPredicateDigest(
		codec.state.current, context.mandatoryPredicate,
	)
	wire := wireClaims{
		Version:                      cursorVersion,
		IssuedAt:                     issuedAt.Unix(),
		ExpiresAt:                    issuedAt.Add(cursorLifetime).Unix(),
		AuthorizationContextSHA256:   encodeDigest(context.authorizationContext),
		ResourceOwner:                owner,
		ResourceID:                   resourceID.String(),
		PublishedVersionID:           publishedVersionID.String(),
		CatalogSHA256:                catalogHex,
		QueryIdentitySHA256:          encodeDigest(context.queryIdentity),
		MandatoryPredicateHMACSHA256: encodeDigest(mandatoryDigest),
		PageSize:                     context.pageSize,
		EffectiveSort:                encodeWireSorts(context.sorts),
		LastValues:                   last,
	}
	if context.scopeKind == "tenant" {
		wire.TenantID = context.scopeID.String()
		wire.SubjectID = context.subjectID.String()
	} else {
		wire.ScopeKind = context.scopeKind
		wire.ScopeID = context.scopeID.String()
		wire.Audience = context.audience
		wire.OperationID = context.operationID
		wire.SnapshotAt = context.snapshotAt.Format(time.RFC3339Nano)
		if context.subjectID != uuid.Nil {
			wire.SubjectID = context.subjectID.String()
		}
	}
	return wire, nil
}
