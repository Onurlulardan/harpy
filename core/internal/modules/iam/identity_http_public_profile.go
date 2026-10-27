package iam

import (
	"context"
	"encoding/json"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func (application *identityHTTPApplication) publicProfileValues(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource identityfacade.ProfileResource,
	document identityfacade.ProfileDocument,
) (map[string]json.RawMessage, error) {
	values, err := identityProfileMap(document.Values())
	if err != nil {
		return nil, err
	}
	names, err := application.identities.profiles.PublicFieldNamesForSchema(
		ctx, scope, resource, document.SchemaID(),
	)
	if err != nil {
		return nil, err
	}
	result := make(map[string]json.RawMessage, len(names))
	for id, name := range names {
		key := uuid.UUID(id).String()
		value, found := values[key]
		if !found {
			continue
		}
		if name == "" {
			return nil, ErrIdentityHTTPUnavailable
		}
		result[name] = value
	}
	return result, nil
}

func (application *identityHTTPApplication) publicProfileDocument(
	ctx context.Context,
	scope platformtx.TenantScope,
	resource identityfacade.ProfileResource,
	document identityfacade.ProfileDocument,
) ([]byte, error) {
	values, err := application.publicProfileValues(ctx, scope, resource, document)
	if err != nil {
		return nil, err
	}
	delete(values, "display_name")
	return identityMarshalPublicProfile(values)
}

func identityMarshalPublicProfile(values map[string]json.RawMessage) ([]byte, error) {
	document, err := json.Marshal(values)
	if err != nil || len(document) > identityfacade.MaximumProfileBytes {
		return nil, ErrIdentityHTTPUnavailable
	}
	return document, nil
}
