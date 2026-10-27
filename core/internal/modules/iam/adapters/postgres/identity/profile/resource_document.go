package profilepostgres

import (
	"context"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

func decodeContractResource(
	ctx context.Context,
	bound *boundStore,
	contract profileWriteContract,
	resource identitydomain.ProfileResource,
	payload []byte,
) (profileWriteResource, error) {
	documentWire, actualSchemaID, digest, err := decodeProfileDocument(payload)
	if err != nil {
		return profileWriteResource{}, identitydomain.ErrStorage
	}
	schema, err := contractResourceSchema(
		ctx, bound, contract, resource, actualSchemaID,
	)
	if err != nil {
		clear(documentWire)
		return profileWriteResource{}, err
	}
	document, err := profileapp.DecodeDocument(
		schema, documentWire, digest[:], profileNormalizer{},
	)
	clear(documentWire)
	if err != nil || document.Resource() != profileapp.ResourceKind(resource) {
		return profileWriteResource{}, identitydomain.ErrStorage
	}
	return profileWriteResource{
		actualSchemaID: actualSchemaID, actualSchema: schema, document: document,
	}, nil
}

func contractResourceSchema(
	ctx context.Context,
	bound *boundStore,
	contract profileWriteContract,
	resource identitydomain.ProfileResource,
	actualSchemaID uuid.UUID,
) (profileapp.Schema, error) {
	if actualSchemaID == contract.source.ID() {
		return contract.source, nil
	}
	if contract.migrating() && actualSchemaID == contract.target.ID() {
		return contract.target, nil
	}
	var payload []byte
	if bound == nil || bound.database.QueryRow(
		ctx, profileHistoricalSQL, actualSchemaID,
	).Scan(&payload) != nil {
		return profileapp.Schema{}, identitydomain.ErrStorage
	}
	schema, err := decodeProfileSchema(payload)
	if err != nil || schema.ID() != actualSchemaID ||
		string(schema.Resource()) != string(resource) ||
		schema.State() != profileapp.SchemaRetired ||
		schema.Version() > contract.source.Version() {
		return profileapp.Schema{}, identitydomain.ErrStorage
	}
	return schema, nil
}
