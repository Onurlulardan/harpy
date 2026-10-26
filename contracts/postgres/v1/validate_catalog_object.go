package postgresv1

import (
	"fmt"
	"regexp"
	"strings"
)

var typeIdentityPattern = regexp.MustCompile(`^[a-z][a-z0-9_ .\[\],]{0,127}$`)

func (object CatalogObject) Validate() error {
	if !validCatalogKind(object.Kind) {
		return fmt.Errorf("catalog object kind is invalid")
	}
	if err := validateDatabaseName(object.DatabaseName, "catalog database"); err != nil {
		return err
	}
	if err := validateIdentifier(object.ObjectName, "catalog object name"); err != nil {
		return err
	}
	withoutSchema := object.Kind == "database" || object.Kind == "schema"
	if withoutSchema != (object.SchemaName == nil) {
		return fmt.Errorf("catalog object schema identity is invalid")
	}
	if object.SchemaName != nil {
		if err := validateIdentifier(*object.SchemaName, "catalog schema"); err != nil {
			return err
		}
	}
	callable := object.Kind == "function" || object.Kind == "procedure"
	if callable != (object.IdentityArguments != nil) {
		return fmt.Errorf("catalog callable identity is invalid")
	}
	if object.IdentityArguments != nil {
		if len(*object.IdentityArguments) > 64 {
			return fmt.Errorf("catalog identity argument count is invalid")
		}
		for _, argument := range *object.IdentityArguments {
			if !typeIdentityPattern.MatchString(argument) {
				return fmt.Errorf("catalog identity argument is invalid")
			}
		}
	}
	if object.Kind == "database" && object.ObjectName != object.DatabaseName {
		return fmt.Errorf("database catalog identity is inconsistent")
	}
	return nil
}

func CatalogObjectKey(object CatalogObject) string {
	schema := ""
	if object.SchemaName != nil {
		schema = *object.SchemaName
	}
	arguments := ""
	if object.IdentityArguments != nil {
		arguments = strings.Join(*object.IdentityArguments, ",")
	}
	return strings.Join([]string{
		object.Kind, object.DatabaseName, schema, object.ObjectName,
		arguments,
	}, "|")
}

func validCatalogKind(value string) bool {
	switch value {
	case "database", "schema", "table", "partitioned_table", "sequence", "view",
		"materialized_view", "function", "procedure", "type", "extension":
		return true
	default:
		return false
	}
}
