package compiler

import (
	"crypto/sha256"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

type statementBinding struct {
	scope         resolved.SnapshotScope
	scopeID       uuid.UUID
	subject       uuid.UUID
	audience      resolved.SnapshotAudience
	operationID   string
	snapshotUnix  int64
	authorization [sha256.Size]byte
	mandatory     [sha256.Size]byte
	catalog       api.CatalogReference
	identity      [sha256.Size]byte
	continuation  [sha256.Size]byte
	validated     bool
}

func newStatementBinding(
	source cursorBinding,
	continuation [sha256.Size]byte,
) (statementBinding, error) {
	scope, scopeErr := source.snapshot.Scope()
	scopeID, scopeIDErr := source.snapshot.ScopeID()
	subject, _, subjectErr := source.snapshot.SubjectID()
	audience, audienceErr := source.snapshot.Audience()
	operationID, operationErr := source.snapshot.OperationID()
	snapshotAt, hasSnapshot, snapshotErr := source.snapshot.SnapshotAt()
	if scopeErr != nil || scopeIDErr != nil || subjectErr != nil ||
		audienceErr != nil || operationErr != nil || snapshotErr != nil ||
		source.authorization == [sha256.Size]byte{} || len(source.mandatory) == 0 ||
		source.reference.Validate() != nil || source.identity == [sha256.Size]byte{} ||
		continuation == [sha256.Size]byte{} {
		return statementBinding{}, ErrContract
	}
	snapshotUnix := int64(0)
	if hasSnapshot {
		snapshotUnix = snapshotAt.UnixMicro()
	}
	binding := statementBinding{
		scope: scope, scopeID: scopeID, subject: subject, audience: audience,
		operationID: operationID, snapshotUnix: snapshotUnix,
		authorization: source.authorization,
		mandatory:     sha256.Sum256(source.mandatory),
		catalog:       source.reference, identity: source.identity,
		continuation: continuation, validated: true,
	}
	if binding.validate() != nil {
		return statementBinding{}, ErrContract
	}
	return binding, nil
}

func (binding statementBinding) validate() error {
	owner, ownerErr := binding.catalog.Owner()
	if !binding.validated || binding.scopeID == uuid.Nil ||
		binding.scopeID.Variant() != uuid.RFC4122 ||
		binding.authorization == [sha256.Size]byte{} ||
		binding.mandatory == [sha256.Size]byte{} || binding.identity == [sha256.Size]byte{} ||
		binding.continuation == [sha256.Size]byte{} ||
		binding.catalog.Validate() != nil || ownerErr != nil || owner == "" {
		return ErrContract
	}
	if binding.scope == resolved.SnapshotScopeTenant {
		if binding.audience != resolved.SnapshotAudienceTenantActor ||
			binding.subject == uuid.Nil || binding.snapshotUnix != 0 ||
			binding.operationID != "restquery.execute" {
			return ErrContract
		}
	} else if binding.scope == resolved.SnapshotScopeGlobal {
		if binding.snapshotUnix <= 0 || binding.operationID == "" {
			return ErrContract
		}
		if binding.audience == resolved.SnapshotAudienceGlobalActor {
			if binding.subject == uuid.Nil {
				return ErrContract
			}
		} else if binding.audience != resolved.SnapshotAudienceGlobalPublic ||
			binding.subject != uuid.Nil {
			return ErrContract
		}
	} else {
		return ErrContract
	}
	return nil
}
