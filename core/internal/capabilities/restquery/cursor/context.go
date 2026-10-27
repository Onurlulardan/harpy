package cursor

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/google/uuid"
)

type Context struct {
	scopeKind            string
	scopeID              uuid.UUID
	subjectID            uuid.UUID
	audience             string
	operationID          string
	snapshotAt           time.Time
	authorizationContext [32]byte
	catalog              api.CatalogReference
	queryIdentity        [32]byte
	mandatoryPredicate   []byte
	pageSize             int
	sorts                []SortDefinition
	valid                bool
}

func NewContext(
	tenantID string,
	subjectID string,
	authorizationContextSHA256 [32]byte,
	catalog api.CatalogReference,
	queryIdentitySHA256 [32]byte,
	mandatoryPredicateCanonical []byte,
	pageSize int,
	sorts []SortDefinition,
) (Context, error) {
	return newContext(contextInput{
		scopeKind:            "tenant",
		scopeID:              tenantID,
		subjectID:            subjectID,
		audience:             "tenant_actor",
		operationID:          "restquery.execute",
		authorizationContext: authorizationContextSHA256,
		catalog:              catalog,
		queryIdentity:        queryIdentitySHA256,
		mandatoryPredicate:   mandatoryPredicateCanonical,
		pageSize:             pageSize,
		sorts:                sorts,
	})
}

func NewGlobalContext(
	environmentID string,
	subjectID string,
	audience string,
	operationID string,
	snapshotAt time.Time,
	authorizationContextSHA256 [32]byte,
	catalog api.CatalogReference,
	queryIdentitySHA256 [32]byte,
	mandatoryPredicateCanonical []byte,
	pageSize int,
	sorts []SortDefinition,
) (Context, error) {
	return newContext(contextInput{
		scopeKind: "global", scopeID: environmentID, subjectID: subjectID,
		audience: audience, operationID: operationID, snapshotAt: snapshotAt,
		authorizationContext: authorizationContextSHA256, catalog: catalog,
		queryIdentity:      queryIdentitySHA256,
		mandatoryPredicate: mandatoryPredicateCanonical,
		pageSize:           pageSize, sorts: sorts,
	})
}

func (context Context) TenantID() (uuid.UUID, error) {
	if err := context.Validate(); err != nil {
		return uuid.Nil, err
	}
	if context.scopeKind != "tenant" {
		return uuid.Nil, errInvalidContext
	}
	return context.scopeID, nil
}

func (context Context) SubjectID() (uuid.UUID, error) {
	if err := context.Validate(); err != nil {
		return uuid.Nil, err
	}
	if context.subjectID == uuid.Nil {
		return uuid.Nil, errInvalidContext
	}
	return context.subjectID, nil
}

func (context Context) ScopeKind() (string, error) {
	if err := context.Validate(); err != nil {
		return "", err
	}
	return context.scopeKind, nil
}

func (context Context) ScopeID() (uuid.UUID, error) {
	if err := context.Validate(); err != nil {
		return uuid.Nil, err
	}
	return context.scopeID, nil
}

func (context Context) Audience() (string, error) {
	if err := context.Validate(); err != nil {
		return "", err
	}
	return context.audience, nil
}

func (context Context) OperationID() (string, error) {
	if err := context.Validate(); err != nil {
		return "", err
	}
	return context.operationID, nil
}

func (context Context) SnapshotAt() (time.Time, bool, error) {
	if err := context.Validate(); err != nil {
		return time.Time{}, false, err
	}
	return context.snapshotAt, !context.snapshotAt.IsZero(), nil
}

func (context Context) AuthorizationContextSHA256() ([32]byte, error) {
	if err := context.Validate(); err != nil {
		return [32]byte{}, err
	}
	return context.authorizationContext, nil
}

func (context Context) Catalog() (api.CatalogReference, error) {
	if err := context.Validate(); err != nil {
		return api.CatalogReference{}, err
	}
	return context.catalog, nil
}

func (context Context) QueryIdentitySHA256() ([32]byte, error) {
	if err := context.Validate(); err != nil {
		return [32]byte{}, err
	}
	return context.queryIdentity, nil
}

func (context Context) MandatoryPredicateCanonical() ([]byte, error) {
	if err := context.Validate(); err != nil {
		return nil, err
	}
	return append([]byte(nil), context.mandatoryPredicate...), nil
}

func (context Context) PageSize() (int, error) {
	if err := context.Validate(); err != nil {
		return 0, err
	}
	return context.pageSize, nil
}

func (context Context) Sorts() ([]SortDefinition, error) {
	if err := context.Validate(); err != nil {
		return nil, err
	}
	return append([]SortDefinition(nil), context.sorts...), nil
}

func (Context) String() string                         { return redacted }
func (Context) GoString() string                       { return redacted }
func (Context) LogValue() slog.Value                   { return redactedLogValue() }
func (context Context) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Context) MarshalJSON() ([]byte, error)           { return rejectJSONMarshal() }
