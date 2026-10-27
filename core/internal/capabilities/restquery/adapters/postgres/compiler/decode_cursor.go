package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func (compiler Compiler) decodeCursor(
	context cursor.Context,
	query resolved.RecordQuery,
) ([]cursor.LastValue, error) {
	page, err := query.Page()
	if err != nil {
		return nil, ErrContract
	}
	token, present, err := page.After()
	if err != nil || !present {
		return nil, err
	}
	if context.Validate() != nil {
		return nil, ErrContract
	}
	claims, err := compiler.codec.Decode(token, context)
	if err != nil {
		return nil, err
	}
	values, err := claims.LastValues()
	if err != nil {
		return nil, ErrContract
	}
	return values, nil
}
