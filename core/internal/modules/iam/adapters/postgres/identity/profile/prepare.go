package profilepostgres

import (
	"context"
	_ "embed"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

//go:embed sql/tenant/active.sql
var profileActiveSQL string

func (storage *store) FieldIDs(
	ctx context.Context,
	scope tenantScope,
	resource identitydomain.ProfileResource,
	keys []string,
) (map[string]identitydomain.IdentityID, error) {
	if !profileAliasRequestValid(ctx, scope, resource, keys) {
		return nil, identitydomain.ErrInvalidCommand
	}
	_, schema, err := storage.activeSchema(ctx, scope, resource)
	if err != nil {
		return nil, identitydomain.ErrStorage
	}
	return profileFieldIDs(schema, keys)
}

func (storage *store) FieldIDsForSchema(
	ctx context.Context,
	scope tenantScope,
	resource identitydomain.ProfileResource,
	schemaID identitydomain.IdentityID,
	keys []string,
) (map[string]identitydomain.IdentityID, error) {
	if !profileAliasRequestValid(ctx, scope, resource, keys) ||
		schemaID == (identitydomain.IdentityID{}) {
		return nil, identitydomain.ErrInvalidCommand
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return nil, identitydomain.ErrStorage
	}
	var payload []byte
	if bound.database.QueryRow(
		ctx, profileHistoricalSQL, databaseID(schemaID),
	).Scan(&payload) != nil {
		return nil, identitydomain.ErrStorage
	}
	schema, err := decodeProfileSchema(payload)
	if err != nil || schema.ID() != databaseID(schemaID) ||
		string(schema.Resource()) != string(resource) {
		return nil, identitydomain.ErrStorage
	}
	return profileFieldIDs(schema, keys)
}

func profileAliasRequestValid(
	ctx context.Context,
	scope tenantScope,
	resource identitydomain.ProfileResource,
	keys []string,
) bool {
	return profileLookupScopeUsable(ctx, scope) && resource.Valid() &&
		len(keys) > 0 && len(keys) <= profileapp.MaximumActiveFields
}

func profileFieldIDs(
	schema profileapp.Schema,
	keys []string,
) (map[string]identitydomain.IdentityID, error) {
	requested := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		if key == "" {
			return nil, identitydomain.ErrInvalidCommand
		}
		requested[key] = struct{}{}
	}
	result := make(map[string]identitydomain.IdentityID, len(requested))
	for _, field := range schema.Fields() {
		if _, found := requested[field.Key()]; found && field.State() == profileapp.FieldActive {
			result[field.Key()] = identitydomain.IdentityID(field.ID())
		}
	}
	if len(result) != len(requested) {
		return nil, identitydomain.ErrInvalidCommand
	}
	return result, nil
}

func profileLookupScopeUsable(ctx context.Context, scope tenantScope) bool {
	return scopeUsable(ctx, scope, commandMode) || scopeUsable(ctx, scope, snapshotMode)
}

func (storage *store) activeSchema(
	ctx context.Context, scope tenantScope, resource identitydomain.ProfileResource,
) (*boundStore, profileapp.Schema, error) {
	bound, err := storage.bind(scope)
	if err != nil {
		return nil, profileapp.Schema{}, err
	}
	var payload []byte
	if bound.database.QueryRow(ctx, profileActiveSQL, string(resource)).Scan(&payload) != nil {
		return nil, profileapp.Schema{}, identitydomain.ErrStorage
	}
	schema, err := decodeProfileSchema(payload)
	if err != nil || schema.State() != profileapp.SchemaActive ||
		string(schema.Resource()) != string(resource) {
		return nil, profileapp.Schema{}, identitydomain.ErrStorage
	}
	return bound, schema, nil
}

func newProfileDocument(
	resource identitydomain.ProfileResource, document profileapp.Document,
) (identitydomain.ProfileDocument, error) {
	return identitydomain.NewProfileDocument(
		resource, identitydomain.IdentityID(document.SchemaID()),
		document.CanonicalJSON(), document.SHA256(),
	)
}
