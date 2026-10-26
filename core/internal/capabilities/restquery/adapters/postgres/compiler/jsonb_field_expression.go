package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func jsonbFieldExpression(
	field api.FieldDefinition,
) (string, bool, error) {
	document, key, present, err := jsonbStorageExpression(field)
	if err != nil || !present {
		return "", present, err
	}
	_, kind, err := fieldIdentity(field)
	if err != nil {
		return "", true, err
	}
	jsonValue := document + " -> " + key
	textValue := document + " ->> " + key
	switch kind {
	case api.LogicalTypeText, api.LogicalTypeEnum:
		return "(" + textValue + ")", true, nil
	case api.LogicalTypeInteger:
		return "CAST(" + textValue + " AS pg_catalog.int8)", true, nil
	case api.LogicalTypeDecimal:
		return "CAST(" + textValue + " AS pg_catalog.numeric)", true, nil
	case api.LogicalTypeBoolean:
		return "CAST(" + textValue + " AS pg_catalog.boolean)", true, nil
	case api.LogicalTypeDate:
		return "CAST(" + textValue + " AS pg_catalog.date)", true, nil
	case api.LogicalTypeTimestamp:
		return "CAST(" + textValue + " AS pg_catalog.timestamptz)", true, nil
	case api.LogicalTypeUUID, api.LogicalTypeRelation:
		return "CAST(" + textValue + " AS pg_catalog.uuid)", true, nil
	case api.LogicalTypeJSON:
		return "(" + jsonValue + ")", true, nil
	case api.LogicalTypeTextList:
		return jsonbListExpression(document, key, false), true, nil
	case api.LogicalTypeUUIDList:
		return jsonbListExpression(document, key, true), true, nil
	default:
		return "", true, ErrContract
	}
}

func jsonbStorageExpression(
	field api.FieldDefinition,
) (string, string, bool, error) {
	storage, err := field.Storage()
	if err != nil {
		return "", "", false, ErrContract
	}
	key, present, err := storage.JSONBKey()
	if err != nil || !present {
		return "", "", present, mapContractError(err)
	}
	column, err := storage.Column()
	if err != nil {
		return "", "", true, ErrContract
	}
	var builder strings.Builder
	if appendIdentifier(&builder, column) != nil {
		return "", "", true, ErrContract
	}
	return builder.String(), "'" + key.String() + "'", true, nil
}

func jsonbListExpression(document, key string, uuidValues bool) string {
	value := "element.value"
	if uuidValues {
		value = "CAST(element.value AS pg_catalog.uuid)"
	}
	return "CASE WHEN " + document + " ? " + key +
		" THEN ARRAY(SELECT " + value +
		" FROM pg_catalog.jsonb_array_elements_text(" + document + " -> " + key +
		") AS element(value)) ELSE NULL END"
}

func mapContractError(err error) error {
	if err != nil {
		return ErrContract
	}
	return nil
}
