package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type cursorBinding struct {
	snapshot      resolved.SnapshotBinding
	authorization [sha256.Size]byte
	mandatory     []byte
	reference     api.CatalogReference
	identity      [sha256.Size]byte
}

func newCursorBinding(
	plan resolved.AuthorizedPlan,
	reference api.CatalogReference,
	identity [sha256.Size]byte,
) (cursorBinding, error) {
	snapshot, snapshotErr := plan.SnapshotBinding()
	authorization, authorizationErr := plan.AuthorizationContextSHA256()
	mandatory, mandatoryErr := plan.MandatoryCanonical()
	if snapshotErr != nil || authorizationErr != nil ||
		mandatoryErr != nil || reference.Validate() != nil ||
		identity == [sha256.Size]byte{} {
		return cursorBinding{}, ErrContract
	}
	return cursorBinding{
		snapshot: snapshot, authorization: authorization,
		mandatory: mandatory, reference: reference, identity: identity,
	}, nil
}

func cursorContext(
	binding cursorBinding,
	query resolved.RecordQuery,
) (cursor.Context, error) {
	page, pageErr := query.Page()
	pageSize, pageSizeErr := page.Size()
	sorts, sortsErr := query.Sorts()
	definitions, definitionErr := cursorSortDefinitions(sorts)
	if pageErr != nil || pageSizeErr != nil || sortsErr != nil || definitionErr != nil {
		return cursor.Context{}, ErrContract
	}
	scope, scopeErr := binding.snapshot.Scope()
	scopeID, scopeIDErr := binding.snapshot.ScopeID()
	subject, hasSubject, subjectErr := binding.snapshot.SubjectID()
	if scopeErr != nil || scopeIDErr != nil || subjectErr != nil {
		return cursor.Context{}, ErrContract
	}
	var context cursor.Context
	var err error
	if scope == resolved.SnapshotScopeTenant {
		if !hasSubject {
			return cursor.Context{}, ErrContract
		}
		context, err = cursor.NewContext(
			scopeID.String(), subject.String(), binding.authorization,
			binding.reference, binding.identity, binding.mandatory, pageSize, definitions,
		)
	} else {
		audience, audienceErr := binding.snapshot.Audience()
		operationID, operationErr := binding.snapshot.OperationID()
		snapshotAt, foundAt, snapshotErr := binding.snapshot.SnapshotAt()
		if audienceErr != nil || operationErr != nil || snapshotErr != nil || !foundAt {
			return cursor.Context{}, ErrContract
		}
		audienceName := "authenticated"
		subjectName := subject.String()
		if audience == resolved.SnapshotAudienceGlobalPublic {
			audienceName, subjectName = "public", ""
		}
		context, err = cursor.NewGlobalContext(
			scopeID.String(), subjectName, audienceName, operationID, snapshotAt,
			binding.authorization, binding.reference, binding.identity,
			binding.mandatory, pageSize, definitions,
		)
	}
	if err != nil {
		return cursor.Context{}, ErrContract
	}
	return context, nil
}

func cursorSortDefinitions(
	sorts []resolved.Sort,
) ([]cursor.SortDefinition, error) {
	definitions := make([]cursor.SortDefinition, len(sorts))
	for index, sort := range sorts {
		field, fieldErr := sort.Field()
		identifier, identifierErr := field.ID()
		fieldType, typeErr := field.Type()
		kind, kindErr := fieldType.Kind()
		capabilities, capabilityErr := field.Capabilities()
		bound, boundErr := capabilities.CursorCanonicalBytes()
		direction, directionErr := sort.Direction()
		if fieldErr != nil || identifierErr != nil || typeErr != nil || kindErr != nil ||
			capabilityErr != nil || boundErr != nil || directionErr != nil {
			return nil, ErrContract
		}
		definition, err := cursor.NewSortDefinition(
			identifier.String(), direction, kind, bound,
		)
		if err != nil {
			return nil, ErrContract
		}
		definitions[index] = definition
	}
	return definitions, nil
}
