package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

func validateCursorContext(
	context cursor.Context,
	binding statementBinding,
	columns []ResultColumn,
	requestedSize int,
) error {
	if context.Validate() != nil || binding.validate() != nil {
		return ErrContract
	}
	scopeKind, scopeKindErr := context.ScopeKind()
	scopeID, scopeIDErr := context.ScopeID()
	actor, hasActor, actorErr := cursorSubject(context)
	audience, audienceErr := context.Audience()
	operationID, operationErr := context.OperationID()
	snapshotAt, hasSnapshot, snapshotErr := context.SnapshotAt()
	authorization, authorizationErr := context.AuthorizationContextSHA256()
	catalog, catalogErr := context.Catalog()
	identity, identityErr := context.QueryIdentitySHA256()
	mandatory, mandatoryErr := context.MandatoryPredicateCanonical()
	pageSize, pageErr := context.PageSize()
	sorts, sortsErr := context.Sorts()
	if scopeKindErr != nil || scopeIDErr != nil || actorErr != nil || audienceErr != nil ||
		operationErr != nil || snapshotErr != nil || authorizationErr != nil ||
		catalogErr != nil || identityErr != nil || mandatoryErr != nil ||
		pageErr != nil || sortsErr != nil || scopeKind != bindingScopeName(binding.scope) ||
		scopeID != binding.scopeID || actor != binding.subject ||
		hasActor != (binding.subject != uuid.Nil) || audience != bindingAudienceName(binding) ||
		operationID != binding.operationID || hasSnapshot != (binding.snapshotUnix > 0) ||
		hasSnapshot && snapshotAt.UnixMicro() != binding.snapshotUnix ||
		authorization != binding.authorization ||
		catalog != binding.catalog || identity != binding.identity ||
		sha256.Sum256(mandatory) != binding.mandatory || pageSize != requestedSize ||
		len(sorts) < 1 || len(sorts) > 6 {
		return ErrContract
	}
	sortedColumns := 0
	for _, column := range columns {
		ordinal, sorted, err := column.SortOrdinal()
		if err != nil || sorted && (ordinal < 0 || ordinal >= len(sorts)) {
			return ErrContract
		}
		if sorted {
			sortedColumns++
		}
	}
	if sortedColumns != len(sorts) {
		return ErrContract
	}
	for ordinal, definition := range sorts {
		if !sortDefinitionMatchesColumn(definition, ordinal, columns) {
			return ErrContract
		}
	}
	return nil
}

func cursorSubject(context cursor.Context) (uuid.UUID, bool, error) {
	subject, err := context.SubjectID()
	if err == nil {
		return subject, true, nil
	}
	audience, audienceErr := context.Audience()
	if audienceErr == nil && audience == "public" {
		return uuid.Nil, false, nil
	}
	return uuid.Nil, false, ErrContract
}

func bindingScopeName(scope resolved.SnapshotScope) string {
	if scope == resolved.SnapshotScopeTenant {
		return "tenant"
	}
	if scope == resolved.SnapshotScopeGlobal {
		return "global"
	}
	return ""
}

func bindingAudienceName(binding statementBinding) string {
	switch binding.audience {
	case resolved.SnapshotAudienceTenantActor:
		return "tenant_actor"
	case resolved.SnapshotAudienceGlobalActor:
		return "authenticated"
	case resolved.SnapshotAudienceGlobalPublic:
		return "public"
	default:
		return ""
	}
}

func sortDefinitionMatchesColumn(
	definition cursor.SortDefinition,
	ordinal int,
	columns []ResultColumn,
) bool {
	fieldID, fieldErr := definition.FieldID()
	logicalType, typeErr := definition.LogicalType()
	direction, directionErr := definition.Direction()
	bound, boundErr := definition.CanonicalBound()
	if fieldErr != nil || typeErr != nil || directionErr != nil || boundErr != nil {
		return false
	}
	matches := 0
	for _, column := range columns {
		columnOrdinal, sorted, ordinalErr := column.SortOrdinal()
		columnID, identifierErr := column.FieldID()
		columnType, columnTypeErr := column.LogicalType()
		columnDirection, columnBound, metadataPresent, metadataErr := column.SortMetadata()
		if ordinalErr != nil || identifierErr != nil || columnTypeErr != nil {
			return false
		}
		if sorted && columnOrdinal == ordinal {
			if metadataErr != nil || !metadataPresent || columnID != fieldID ||
				columnType != logicalType || columnDirection != direction || columnBound != bound {
				return false
			}
			matches++
		}
	}
	return matches == 1
}
