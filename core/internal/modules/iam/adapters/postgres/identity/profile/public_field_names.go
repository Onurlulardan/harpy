package profilepostgres

import (
	"context"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

func (storage *store) PublicFieldNamesForSchema(
	ctx context.Context,
	scope tenantScope,
	resource identitydomain.ProfileResource,
	schemaID identitydomain.IdentityID,
) (map[identitydomain.IdentityID]string, error) {
	if !profileLookupScopeUsable(ctx, scope) || !resource.Valid() ||
		schemaID == (identitydomain.IdentityID{}) {
		return nil, identitydomain.ErrInvalidQuery
	}
	bound, err := storage.bind(scope)
	if err != nil {
		return nil, identitydomain.ErrStorage
	}
	var payload []byte
	if bound.database.QueryRow(ctx, profileHistoricalSQL, databaseID(schemaID)).
		Scan(&payload) != nil {
		return nil, identitydomain.ErrStorage
	}
	schema, err := decodeProfileSchema(payload)
	if err != nil || schema.ID() != databaseID(schemaID) ||
		string(schema.Resource()) != string(resource) {
		return nil, identitydomain.ErrStorage
	}
	result := make(map[identitydomain.IdentityID]string)
	for _, field := range schema.Fields() {
		if field.State() == profileapp.FieldActive &&
			field.Classification() == profileapp.ClassificationPublic {
			result[identitydomain.IdentityID(field.ID())] = field.Key()
		}
	}
	return result, nil
}
