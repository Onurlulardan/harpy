package profile

import (
	"encoding/hex"

	restquery "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

type queryDefinition struct {
	Contract   string                 `json:"contract"`
	SchemaID   string                 `json:"schema_id"`
	SchemaHash string                 `json:"schema_sha256"`
	Resource   ResourceKind           `json:"resource_kind"`
	Version    int64                  `json:"version"`
	Fields     []queryFieldDefinition `json:"fields"`
}

type queryFieldDefinition struct {
	ID         string   `json:"id"`
	PublicName string   `json:"public_name"`
	Type       string   `json:"logical_type"`
	Operators  []string `json:"operators"`
	Selectable bool     `json:"selectable"`
	Sortable   bool     `json:"sortable"`
	Searchable bool     `json:"searchable"`
}

func buildQueryDefinition(schema Schema, fields []QueryField) queryDefinition {
	result := queryDefinition{
		Contract: "harpy/iam-profile-query-catalog/v1",
		SchemaID: schema.id.String(), SchemaHash: hex.EncodeToString(schema.definitionSHA[:]),
		Resource: schema.resource, Version: schema.version,
		Fields: make([]queryFieldDefinition, 0, len(fields)),
	}
	for _, field := range fields {
		result.Fields = append(result.Fields, queryFieldDefinition{
			ID: field.id.String(), PublicName: field.publicName,
			Type: field.logicalType.String(), Operators: operatorNames(field.operators),
			Selectable: true,
		})
	}
	return result
}

func operatorNames(operators []restquery.Operator) []string {
	result := make([]string, len(operators))
	for index, operator := range operators {
		result[index] = operator.String()
	}
	return result
}
