package cursor

import "time"

type bindingResult uint8

const (
	bindingInvalid bindingResult = iota
	bindingValid
	bindingStale
)

func bindExpected(
	claims decodedClaims,
	expected Context,
	key codecKey,
) bindingResult {
	expectedOwner, ownerErr := expected.catalog.Owner()
	expectedResource, resourceErr := expected.catalog.ResourceID()
	expectedVersion, versionErr := expected.catalog.PublishedVersionID()
	expectedHash, hashErr := expected.catalog.CatalogHash()
	claimOwner, claimOwnerErr := claims.reference.Owner()
	claimResource, claimResourceErr := claims.reference.ResourceID()
	claimVersion, claimVersionErr := claims.reference.PublishedVersionID()
	claimHash, claimHashErr := claims.reference.CatalogHash()
	expectedHashBytes, expectedHashErr := expectedHash.Bytes()
	claimHashBytes, claimHashBytesErr := claimHash.Bytes()
	if ownerErr != nil || resourceErr != nil || versionErr != nil || hashErr != nil ||
		claimOwnerErr != nil || claimResourceErr != nil || claimVersionErr != nil ||
		claimHashErr != nil || expectedHashErr != nil || claimHashBytesErr != nil {
		return bindingInvalid
	}
	mandatory := mandatoryPredicateDigest(key, expected.mandatoryPredicate)
	contextMatch := constantTimeString(claims.scopeKind, expected.scopeKind)
	contextMatch &= constantTimeBytes(claims.scopeID[:], expected.scopeID[:])
	contextMatch &= constantTimeBytes(claims.subjectID[:], expected.subjectID[:])
	contextMatch &= constantTimeString(claims.audience, expected.audience)
	contextMatch &= constantTimeString(claims.operationID, expected.operationID)
	contextMatch &= constantTimeString(
		claims.snapshotAt.Format(time.RFC3339Nano),
		expected.snapshotAt.Format(time.RFC3339Nano),
	)
	contextMatch &= constantTimeBytes(
		claims.authorizationContext[:], expected.authorizationContext[:],
	)
	contextMatch &= constantTimeBytes(claims.queryIdentity[:], expected.queryIdentity[:])
	contextMatch &= constantTimeBytes(claims.mandatoryDigest[:], mandatory[:])
	contextMatch &= constantTimeInt(claims.pageSize, expected.pageSize)
	contextMatch &= constantTimeSorts(claims.sorts, expected.sorts)
	ownerResourceMatch := constantTimeString(claimOwner, expectedOwner)
	ownerResourceMatch &= constantTimeBytes(claimResource[:], expectedResource[:])
	if contextMatch != 1 || ownerResourceMatch != 1 ||
		validateLastValues(expected.sorts, claims.last) != nil {
		return bindingInvalid
	}
	versionMatch := constantTimeBytes(claimVersion[:], expectedVersion[:])
	hashMatch := constantTimeBytes(claimHashBytes, expectedHashBytes)
	if versionMatch != 1 || hashMatch != 1 {
		return bindingStale
	}
	return bindingValid
}
