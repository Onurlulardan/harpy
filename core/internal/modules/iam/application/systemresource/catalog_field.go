package systemresource

import (
	restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
	formpresentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
)

func presentationField(
	published catalog.Published,
	resource publicresource.Resource,
	definition restquery.FieldDefinition,
	visibility CatalogVisibility,
) (formpresentation.Field, bool, error) {
	fieldID, idErr := definition.ID()
	fieldType, typeErr := definition.Type()
	logicalType, kindErr := fieldType.Kind()
	classification, classErr := published.Classification(fieldID)
	if idErr != nil || typeErr != nil || kindErr != nil || classErr != nil {
		return formpresentation.Field{}, false, publicresource.ErrContract
	}
	if !visibility.allows(classification) {
		return formpresentation.Field{}, false, nil
	}
	typeID := fieldregistry.TypeID(logicalType.String())
	valueProfile, valid := formpresentation.DefaultValueProfile(typeID)
	if !valid {
		return formpresentation.Field{}, false, publicresource.ErrContract
	}
	field, err := formpresentation.NewField(
		fieldID.String(), typeID, valueProfile, fieldMode(resource, fieldID.String()),
	)
	if err != nil {
		return formpresentation.Field{}, false, publicresource.ErrContract
	}
	return field, true, nil
}
