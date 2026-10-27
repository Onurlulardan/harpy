package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func appendResourceStorage(builder *strings.Builder, storage api.ResourceStorage) error {
	if builder == nil || storage.Validate() != nil {
		return ErrContract
	}
	schema, qualified, err := storage.Schema()
	if err != nil {
		return ErrContract
	}
	if qualified {
		if appendIdentifier(builder, schema) != nil {
			return ErrContract
		}
		builder.WriteByte('.')
	}
	table, err := storage.Table()
	if err != nil || appendIdentifier(builder, table) != nil {
		return ErrContract
	}
	return nil
}
