package iam

import (
	"context"
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (port *identitySystemResourcePort) resolveRelationSummaries(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence identityPublicEvidence,
	resource publicresource.Resource,
	ids []uuid.UUID,
) (map[uuid.UUID]publicresource.RelationSummary, bool, error) {
	authority, err := newIdentitySystemResourceAuthorization(
		port.application, ctx, scope, evidence, resource,
	)
	if err != nil {
		return nil, false, err
	}
	if _, err = authority.require("read"); err != nil {
		if relationReadForbidden(err) {
			return map[uuid.UUID]publicresource.RelationSummary{}, false, nil
		}
		return nil, false, err
	}
	catalog, published, err := port.resolveCatalogSnapshot(
		ctx, scope, resource, authority,
	)
	if err != nil {
		return nil, false, err
	}
	fields, err := identityRelationTargetFields(resource, catalog)
	if err != nil {
		return nil, false, err
	}
	result := make(map[uuid.UUID]publicresource.RelationSummary, len(ids))
	for start := 0; start < len(ids); start += identityRelationQueryBatchSize {
		end := min(start+identityRelationQueryBatchSize, len(ids))
		request, requestErr := newIdentityRelationQuery(ids[start:end], fields)
		if requestErr != nil {
			return nil, false, requestErr
		}
		query, queryErr := port.application.queries.RunValue(
			ctx, scope, published, request,
		)
		if queryErr != nil || query.HasMore() || query.NextCursor() != "" {
			return nil, false, publicresource.ErrUnavailable
		}
		batch, parseErr := identityRelationSummaries(
			resource, query.Items(), fields,
		)
		if parseErr != nil || len(batch) != end-start {
			return nil, false, publicresource.ErrUnavailable
		}
		for id, summary := range batch {
			if _, duplicate := result[id]; duplicate {
				return nil, false, publicresource.ErrUnavailable
			}
			result[id] = summary
		}
	}
	return result, true, nil
}

func identityRelationTargetFields(
	resource publicresource.Resource,
	catalog publicresource.FieldCatalog,
) ([]string, error) {
	descriptors, err := catalog.Descriptors()
	if err != nil {
		return nil, publicresource.ErrUnavailable
	}
	available := make(map[string]struct{}, len(descriptors))
	for _, descriptor := range descriptors {
		name, nameErr := descriptor.PublicName()
		if nameErr != nil {
			return nil, publicresource.ErrUnavailable
		}
		available[name] = struct{}{}
	}
	required := "name"
	candidates := []string{"name", "slug", "status"}
	if resource == publicresource.Users {
		required = "display_name"
		candidates = []string{"display_name", "email", "status"}
	}
	if _, found := available["id"]; !found {
		return nil, publicresource.ErrUnavailable
	}
	if _, found := available[required]; !found {
		if resource != publicresource.Users {
			return nil, publicresource.ErrUnavailable
		}
		if _, emailFound := available["email"]; !emailFound {
			return nil, publicresource.ErrUnavailable
		}
	}
	result := []string{"id"}
	for _, candidate := range candidates {
		if _, found := available[candidate]; found {
			result = append(result, candidate)
		}
	}
	return result, nil
}

func identityRelationSummaries(
	resource publicresource.Resource,
	items []json.RawMessage,
	fields []string,
) (map[uuid.UUID]publicresource.RelationSummary, error) {
	allowed := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		allowed[field] = struct{}{}
	}
	result := make(map[uuid.UUID]publicresource.RelationSummary, len(items))
	for _, item := range items {
		id, summary, err := identityRelationSummary(resource, item, allowed)
		if err != nil {
			return nil, err
		}
		if _, duplicate := result[id]; duplicate {
			return nil, publicresource.ErrUnavailable
		}
		result[id] = summary
	}
	return result, nil
}
