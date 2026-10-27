package iam

import (
	"context"
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (port *identitySystemResourcePort) relationReadItem(
	ctx context.Context,
	scope platformtx.TenantScope,
	published catalog.Published,
	resource publicresource.Resource,
	fieldCatalog publicresource.FieldCatalog,
	id uuid.UUID,
) (json.RawMessage, error) {
	fields, err := identityRelationSourceFields(resource, fieldCatalog)
	if err != nil {
		return nil, err
	}
	request, err := newIdentityRelationQuery([]uuid.UUID{id}, fields)
	if err != nil {
		return nil, err
	}
	result, err := port.application.queries.RunValue(
		ctx, scope, published, request,
	)
	if err != nil || result.HasMore() || result.NextCursor() != "" {
		return nil, publicresource.ErrUnavailable
	}
	items := result.Items()
	if len(items) != 1 {
		return nil, publicresource.ErrUnavailable
	}
	var document map[string]json.RawMessage
	if json.Unmarshal(items[0], &document) != nil {
		return nil, publicresource.ErrUnavailable
	}
	returnedID, err := identityRelationDocumentID(document["id"])
	if err != nil || returnedID != id {
		return nil, publicresource.ErrUnavailable
	}
	return items[0], nil
}
