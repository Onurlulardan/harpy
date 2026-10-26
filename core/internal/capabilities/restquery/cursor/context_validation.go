package cursor

import (
	"bytes"
	"regexp"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/google/uuid"
)

var globalOperationPattern = regexp.MustCompile(`^[a-z][a-z0-9_.:-]{0,127}$`)

type contextInput struct {
	scopeKind            string
	scopeID              string
	subjectID            string
	audience             string
	operationID          string
	snapshotAt           time.Time
	authorizationContext [32]byte
	catalog              api.CatalogReference
	queryIdentity        [32]byte
	mandatoryPredicate   []byte
	pageSize             int
	sorts                []SortDefinition
}

func newContext(input contextInput) (Context, error) {
	scopeID, scopeOK := parseCanonicalUUID(input.scopeID)
	subject, subjectOK := uuid.Nil, input.subjectID == ""
	if input.subjectID != "" {
		subject, subjectOK = parseCanonicalUUID(input.subjectID)
	}
	canonical, canonicalErr := canonicalMandatoryPredicate(input.mandatoryPredicate)
	if !scopeOK || !subjectOK || canonicalErr != nil || input.catalog.Validate() != nil {
		return Context{}, errInvalidContext
	}
	result := Context{
		scopeKind:            input.scopeKind,
		scopeID:              scopeID,
		subjectID:            subject,
		audience:             input.audience,
		operationID:          input.operationID,
		snapshotAt:           input.snapshotAt.UTC().Truncate(time.Second),
		authorizationContext: input.authorizationContext,
		catalog:              input.catalog,
		queryIdentity:        input.queryIdentity,
		mandatoryPredicate:   canonical,
		pageSize:             input.pageSize,
		sorts:                append([]SortDefinition(nil), input.sorts...),
		valid:                true,
	}
	if err := result.Validate(); err != nil {
		return Context{}, err
	}
	return result, nil
}

func canonicalMandatoryPredicate(document []byte) ([]byte, error) {
	if len(document) < 1 || len(document) > maximumMandatoryBytes {
		return nil, errInvalidContext
	}
	canonical, err := attestation.Canonicalize(document)
	if err != nil || len(canonical) < 1 || len(canonical) > maximumMandatoryBytes {
		return nil, errInvalidContext
	}
	return canonical, nil
}

func (context Context) Validate() error {
	if !context.valid || context.scopeID == uuid.Nil ||
		context.catalog.Validate() != nil || context.authorizationContext == [32]byte{} ||
		context.queryIdentity == [32]byte{} || context.pageSize < 1 ||
		context.pageSize > maximumPageSize || len(context.mandatoryPredicate) < 1 ||
		len(context.mandatoryPredicate) > maximumMandatoryBytes || !validSorts(context.sorts) {
		return errInvalidContext
	}
	if context.scopeKind == "tenant" {
		if context.audience != "tenant_actor" || context.subjectID == uuid.Nil ||
			context.operationID != "restquery.execute" || !context.snapshotAt.IsZero() {
			return errInvalidContext
		}
	} else if context.scopeKind == "global" {
		if context.snapshotAt.IsZero() || context.snapshotAt.Location() != time.UTC ||
			!globalOperationPattern.MatchString(context.operationID) {
			return errInvalidContext
		}
		if context.audience == "authenticated" {
			if context.subjectID == uuid.Nil {
				return errInvalidContext
			}
		} else if context.audience != "public" || context.subjectID != uuid.Nil {
			return errInvalidContext
		}
	} else {
		return errInvalidContext
	}
	canonical, err := attestation.Canonicalize(context.mandatoryPredicate)
	defer clear(canonical)
	if err != nil || !bytes.Equal(canonical, context.mandatoryPredicate) {
		return errInvalidContext
	}
	return nil
}

func validSorts(sorts []SortDefinition) bool {
	if len(sorts) < 1 || len(sorts) > maximumSorts {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(sorts))
	for _, sort := range sorts {
		if sort.Validate() != nil {
			return false
		}
		if _, exists := seen[sort.fieldID]; exists {
			return false
		}
		seen[sort.fieldID] = struct{}{}
	}
	return true
}
