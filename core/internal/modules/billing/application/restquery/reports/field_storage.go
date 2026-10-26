package reports

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func resourceStorage(spec resourceSpec) (api.ResourceStorage, error) {
	schema, schemaErr := sqlidentifier.MintStatic("platform")
	table, tableErr := sqlidentifier.MintStatic(spec.View)
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
	storage, err := api.NewFieldStorage(column, nil, nil)
	if err != nil {
		return api.FieldStorage{}, ErrContract
	}
	return storage, nil
}
