package discount

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func resourceStorage() (api.ResourceStorage, error) {
	schema, schemaErr := sqlidentifier.MintStatic("platform")
	table, tableErr := sqlidentifier.MintStatic("billing_discount_query_v1")
	if schemaErr != nil || tableErr != nil {
		return api.ResourceStorage{}, ErrContract
	}
	storage, err := api.NewQualifiedResourceStorage(schema, table)
	if err != nil {
		return api.ResourceStorage{}, ErrContract
	}
	return storage, nil
}

func fieldStorage(spec fieldSpec) (api.FieldStorage, error) {
	column, err := sqlidentifier.MintStatic(spec.Column)
	if err != nil {
		return api.FieldStorage{}, ErrContract
	}
	var folded *sqlidentifier.Token
	if spec.Folded != "" {
		value, foldedErr := sqlidentifier.MintStatic(spec.Folded)
		if foldedErr != nil {
			return api.FieldStorage{}, ErrContract
		}
		folded = &value
	}
	storage, err := api.NewFieldStorage(column, folded, nil)
	if err != nil {
		return api.FieldStorage{}, ErrContract
	}
	return storage, nil
}
