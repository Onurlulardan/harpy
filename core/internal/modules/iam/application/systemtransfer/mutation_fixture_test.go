package systemtransfer

import (
	"crypto/sha256"
	"encoding/json"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
	formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
)

func fixtureMutationResult(
	resource publicresource.Resource,
	recordID publicresource.ID,
	revision int64,
	replayed bool,
) (publicresource.MutationResult, error) {
	return fixtureMutationResultVersions(
		resource, recordID, revision, revision, replayed,
	)
}

func fixtureMutationResultVersions(
	resource publicresource.Resource,
	recordID publicresource.ID,
	recordRevision int64,
	receiptRevision int64,
	replayed bool,
) (publicresource.MutationResult, error) {
	fieldID, name := fixtureIdentityField(resource)
	presentationField, err := formpresentation.NewField(
		fieldID.String(), fieldregistry.TypeUUID,
		fieldregistry.ValueUUIDRFC4122LowerV1, formpresentation.FieldReadOnly,
	)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	presentation, err := formpresentation.NewFieldCatalog(
		[]formpresentation.Field{presentationField},
	)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	descriptor, err := publicresource.NewFieldDescriptor(fieldID, name)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	resourceIDText, _ := resource.CatalogResourceID()
	resourceID, _ := publicresource.ParseID(resourceIDText)
	versionID, _ := publicresource.ParseID("018f0f4e-7b61-7c25-8000-000000000031")
	catalog, err := publicresource.NewFieldCatalog(resource, resourceID, versionID,
		sha256.Sum256([]byte("query-catalog")), presentation,
		[]publicresource.FieldDescriptor{descriptor})
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	values, err := publicresource.NewFieldValues(map[publicresource.ID]json.RawMessage{
		fieldID: json.RawMessage(`"` + recordID.String() + `"`),
	})
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	record, err := publicresource.NewRecord(resource, recordID, recordRevision, values)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	read, err := publicresource.NewReadResult(record, catalog)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	operationID, _ := publicresource.ParseID("018f0f4e-7b61-7c25-8000-000000000021")
	receipt, err := publicresource.NewMutationReceipt(operationID,
		sha256.Sum256([]byte("request")), receiptRevision, replayed)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	return publicresource.NewMutationResult(read, receipt)
}

func fixtureIdentityField(resource publicresource.Resource) (publicresource.ID, string) {
	value, name := publicresource.UserIDFieldID, "id"
	if resource == publicresource.Roles {
		value, name = publicresource.RoleIDFieldID, "id"
	}
	if resource == publicresource.Organizations {
		value, name = publicresource.OrganizationIDFieldID, "id"
	}
	id, _ := publicresource.ParseID(value)
	return id, name
}
