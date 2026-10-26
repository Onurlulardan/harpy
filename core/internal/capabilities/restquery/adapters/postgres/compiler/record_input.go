package compiler

import (
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type recordInput struct {
	storage    api.ResourceStorage
	query      resolved.RecordQuery
	visibility resolved.ResourceVisibility
	last       []cursor.LastValue
	context    cursor.Context
	binding    statementBinding
}

func (input recordInput) validate() error {
	if input.storage.Validate() != nil || input.query.Validate() != nil ||
		input.visibility.Validate() != nil || input.context.Validate() != nil ||
		input.binding.validate() != nil || validateRecordPageShape(input.query) != nil {
		return ErrContract
	}
	for _, value := range input.last {
		if value.Validate() != nil {
			return ErrContract
		}
	}
	return nil
}
