package systemresource

import (
	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
	formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

type recordProjectionField struct {
	id     string
	name   string
	typeID fieldregistry.TypeID
}

func appendRecordProjectionCatalog(
	resource publicresource.Resource,
	fields []formpresentation.Field,
	descriptors []publicresource.FieldDescriptor,
) ([]formpresentation.Field, []publicresource.FieldDescriptor, error) {
	for _, projection := range recordProjectionFields(resource) {
		profile, valid := formpresentation.DefaultValueProfile(projection.typeID)
		if !valid {
			return nil, nil, publicresource.ErrContract
		}
		field, err := formpresentation.NewField(
			projection.id, projection.typeID, profile, formpresentation.FieldReadOnly,
		)
		if err != nil {
			return nil, nil, publicresource.ErrContract
		}
		fieldID, err := publicresource.ParseID(projection.id)
		if err != nil {
			return nil, nil, publicresource.ErrContract
		}
		descriptor, err := publicresource.NewFieldDescriptor(fieldID, projection.name)
		if err != nil {
			return nil, nil, publicresource.ErrContract
		}
		fields = append(fields, field)
		descriptors = append(descriptors, descriptor)
	}
	return fields, descriptors, nil
}

func recordProjectionFields(resource publicresource.Resource) []recordProjectionField {
	switch resource {
	case publicresource.Users:
		return []recordProjectionField{
			{id: publicresource.UserEmailVerifiedFieldID,
				name: "email_verified", typeID: fieldregistry.TypeBoolean},
			{id: publicresource.UserAccountTypeFieldID,
				name: "account_type", typeID: fieldregistry.TypeEnum},
			{id: publicresource.UserOrganizationIDFieldID,
				name: "organization_id", typeID: fieldregistry.TypeRelation},
		}
	case publicresource.Organizations:
		return []recordProjectionField{
			{id: publicresource.OrganizationOwnerIDFieldID,
				name: "owner_id", typeID: fieldregistry.TypeRelation},
		}
	default:
		return nil
	}
}
