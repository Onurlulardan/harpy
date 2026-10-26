package catalog

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

func customFields(
	resource profile.ResourceKind,
	queryFields []profile.QueryField,
	existing map[uuid.UUID]profile.Classification,
) ([]api.FieldDefinition, []uuid.UUID, error) {
	document, err := sqlidentifier.MintStatic("profile_values")
	if err != nil {
		return nil, nil, ErrContract
	}
	definitions := make([]api.FieldDefinition, len(queryFields))
	search := make([]uuid.UUID, 0, 1)
	for index, field := range queryFields {
		identifier, idErr := field.ID()
		name, nameErr := field.PublicName()
		fieldType, typeErr := field.FieldType()
		classification, classErr := field.Classification()
		operators, operatorsErr := field.Operators()
		if idErr != nil || nameErr != nil || typeErr != nil || classErr != nil ||
			operatorsErr != nil || !classification.Valid() {
			return nil, nil, ErrContract
		}
		if _, duplicate := existing[identifier]; duplicate {
			return nil, nil, ErrContract
		}
		storage, finalType, searchable, storageErr := customFieldContract(
			resource, name, classification, fieldType, document, identifier,
		)
		if storageErr != nil {
			return nil, nil, ErrContract
		}
		operatorValues, operatorErr := operatorNames(operators)
		capabilities, capabilitiesErr := fieldCapabilities(
			finalType, operatorValues, false, 0, searchable,
		)
		if storageErr != nil || operatorErr != nil || capabilitiesErr != nil {
			return nil, nil, ErrContract
		}
		definition, definitionErr := api.NewFieldDefinition(
			identifier.String(), name, finalType, storage, capabilities,
		)
		if definitionErr != nil {
			return nil, nil, ErrContract
		}
		definitions[index] = definition
		existing[identifier] = classification
		if searchable {
			search = append(search, identifier)
		}
	}
	if resource == profile.ResourceUser && len(search) != 1 {
		return nil, nil, ErrContract
	}
	return definitions, search, nil
}

func customFieldContract(
	resource profile.ResourceKind,
	name string,
	classification profile.Classification,
	fieldType api.FieldType,
	document sqlidentifier.Token,
	identifier uuid.UUID,
) (api.FieldStorage, api.FieldType, bool, error) {
	if resource != profile.ResourceUser || name != "display_name" {
		storage, err := api.NewJSONBFieldStorage(document, identifier.String())
		return storage, fieldType, false, err
	}
	kind, err := fieldType.Kind()
	if err != nil || kind != api.LogicalTypeText ||
		classification != profile.ClassificationPublic {
		return api.FieldStorage{}, api.FieldType{}, false, ErrContract
	}
	text, err := api.NewTextProfile(
		api.TextCasefoldUnicodeDefaultV1, api.TextAccentDisabled,
	)
	if err != nil {
		return api.FieldStorage{}, api.FieldType{}, false, ErrContract
	}
	searchType, err := api.NewTextFieldType(text)
	if err != nil {
		return api.FieldStorage{}, api.FieldType{}, false, ErrContract
	}
	column, columnErr := sqlidentifier.MintStatic("display_name")
	folded, foldedErr := sqlidentifier.MintStatic("display_name_folded")
	if columnErr != nil || foldedErr != nil {
		return api.FieldStorage{}, api.FieldType{}, false, ErrContract
	}
	storage, err := api.NewFieldStorage(column, &folded, nil)
	return storage, searchType, true, err
}
