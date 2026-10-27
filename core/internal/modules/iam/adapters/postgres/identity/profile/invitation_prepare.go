package profilepostgres

import (
	"encoding/json"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

func PrepareInvitationUserProfile(
	catalog []byte,
	firstName, lastName string,
	custom []byte,
) (identitydomain.ProfileDocument, error) {
	if len(catalog) == 0 || firstName == "" || lastName == "" {
		return identitydomain.ProfileDocument{}, identitydomain.ErrInvalidCommand
	}
	contract, err := invitationWriteContract(catalog)
	if err != nil || contract.source.State() != profileapp.SchemaActive ||
		contract.source.Resource() != profileapp.ResourceUser {
		return identitydomain.ProfileDocument{}, identitydomain.ErrStorage
	}
	values, err := invitationProfileValues(custom)
	if err != nil {
		return identitydomain.ProfileDocument{}, identitydomain.ErrInvalidCommand
	}
	for key, value := range map[string]string{
		"first_name": firstName, "last_name": lastName,
	} {
		fieldID, found := invitationProfileField(contract.source, key)
		if !found {
			return identitydomain.ProfileDocument{}, identitydomain.ErrInvalidCommand
		}
		if _, collision := values[fieldID]; collision {
			return identitydomain.ProfileDocument{}, identitydomain.ErrInvalidCommand
		}
		values[fieldID], _ = json.Marshal(value)
	}
	wire, err := json.Marshal(values)
	if err != nil {
		return identitydomain.ProfileDocument{}, identitydomain.ErrInvalidCommand
	}
	document, err := profileapp.DecodeDocument(
		contract.source, wire, nil, profileNormalizer{},
	)
	clear(wire)
	if err != nil {
		return identitydomain.ProfileDocument{}, identitydomain.ErrInvalidCommand
	}
	if contract.migrating() {
		document, err = migrateProfileWrite(contract, document)
		if err != nil {
			return identitydomain.ProfileDocument{}, identitydomain.ErrInvalidCommand
		}
	}
	return newProfileDocument(identitydomain.ProfileUser, document)
}

func invitationWriteContract(catalog []byte) (profileWriteContract, error) {
	contract, err := decodeProfileWriteContract(catalog)
	if err == nil {
		return contract, nil
	}
	schema, schemaErr := decodeProfileSchema(catalog)
	if schemaErr != nil {
		return profileWriteContract{}, errProfile
	}
	return profileWriteContract{source: schema}, nil
}

func invitationProfileValues(raw []byte) (map[uuid.UUID]json.RawMessage, error) {
	if len(raw) == 0 {
		return make(map[uuid.UUID]json.RawMessage), nil
	}
	var wire map[string]json.RawMessage
	if json.Unmarshal(raw, &wire) != nil || wire == nil || len(wire) > 100 {
		return nil, identitydomain.ErrInvalidCommand
	}
	values := make(map[uuid.UUID]json.RawMessage, len(wire))
	for key, value := range wire {
		id, err := uuid.Parse(key)
		if err != nil || id == uuid.Nil || id.String() != key || len(value) == 0 {
			return nil, identitydomain.ErrInvalidCommand
		}
		values[id] = value
	}
	return values, nil
}

func invitationProfileField(schema profileapp.Schema, key string) (uuid.UUID, bool) {
	for _, field := range schema.Fields() {
		if field.Key() == key && field.State() == profileapp.FieldActive {
			return field.ID(), true
		}
	}
	return uuid.Nil, false
}
