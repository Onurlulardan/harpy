package compiler

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"

func writeCursorContext(writer *sealWriter, context cursor.Context) error {
	if writer == nil || context.Validate() != nil {
		return ErrContract
	}
	scopeKind, scopeKindErr := context.ScopeKind()
	scopeID, scopeIDErr := context.ScopeID()
	actor, _, actorErr := cursorSubject(context)
	audience, audienceErr := context.Audience()
	operationID, operationErr := context.OperationID()
	snapshotAt, hasSnapshot, snapshotErr := context.SnapshotAt()
	authorization, authorizationErr := context.AuthorizationContextSHA256()
	catalog, catalogErr := context.Catalog()
	identity, identityErr := context.QueryIdentitySHA256()
	mandatory, mandatoryErr := context.MandatoryPredicateCanonical()
	pageSize, pageErr := context.PageSize()
	sorts, sortsErr := context.Sorts()
	if scopeKindErr != nil || scopeIDErr != nil || actorErr != nil ||
		audienceErr != nil || operationErr != nil || snapshotErr != nil ||
		authorizationErr != nil ||
		catalogErr != nil || identityErr != nil || mandatoryErr != nil ||
		pageErr != nil || sortsErr != nil {
		return ErrContract
	}
	writer.writeString(scopeKind)
	writer.writeBytes(scopeID[:])
	writer.writeBytes(actor[:])
	writer.writeString(audience)
	writer.writeString(operationID)
	if hasSnapshot {
		writer.writeInt(snapshotAt.UnixMicro())
	} else {
		writer.writeInt(0)
	}
	writer.writeBytes(authorization[:])
	if writeCatalogReference(writer, catalog) != nil {
		return ErrContract
	}
	writer.writeBytes(identity[:])
	writer.writeBytes(mandatory)
	writer.writeInt(int64(pageSize))
	writer.writeInt(int64(len(sorts)))
	for _, definition := range sorts {
		fieldID, fieldErr := definition.FieldID()
		direction, directionErr := definition.Direction()
		logicalType, typeErr := definition.LogicalType()
		bound, boundErr := definition.CanonicalBound()
		if fieldErr != nil || directionErr != nil || typeErr != nil || boundErr != nil {
			return ErrContract
		}
		writer.writeBytes(fieldID[:])
		writer.writeInt(int64(direction))
		writer.writeInt(int64(logicalType))
		writer.writeInt(int64(bound))
	}
	return nil
}
