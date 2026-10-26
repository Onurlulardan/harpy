package restquery

import (
	"context"

	"github.com/google/uuid"
)

type evidenceContextKey struct{}

type Evidence struct {
	actorID, tenantID                  uuid.UUID
	tokenVersion, authorizationVersion int64
	resource, operation                string
	ids                                map[string]string
}

func NewEvidence(actorID, tenantID uuid.UUID, tokenVersion, authorizationVersion int64,
	resource, operation string, ids map[string]string,
) (Evidence, error) {
	value := Evidence{actorID: actorID, tenantID: tenantID, tokenVersion: tokenVersion,
		authorizationVersion: authorizationVersion, resource: resource,
		operation: operation, ids: cloneIDs(ids)}
	if !value.valid() {
		return Evidence{}, ErrInvalid
	}
	return value, nil
}

func (evidence Evidence) valid() bool {
	return evidence.actorID != uuid.Nil && evidence.tenantID != uuid.Nil &&
		evidence.tokenVersion > 0 && evidence.authorizationVersion > 0 &&
		evidence.operation == operationFor(evidence.resource)
}

func cloneIDs(ids map[string]string) map[string]string {
	result := make(map[string]string, len(ids))
	for key, value := range ids {
		result[key] = value
	}
	return result
}

func withEvidence(ctx context.Context, evidence Evidence) context.Context {
	return context.WithValue(ctx, evidenceContextKey{}, evidence)
}

func evidenceFrom(ctx context.Context) (Evidence, bool) {
	if ctx == nil {
		return Evidence{}, false
	}
	value, ok := ctx.Value(evidenceContextKey{}).(Evidence)
	return value, ok && value.valid()
}
