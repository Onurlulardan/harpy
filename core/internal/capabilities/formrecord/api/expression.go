package recordkernel

import (
	"github.com/google/uuid"
)

type expression struct {
	kind      string
	typeID    TypeID
	fieldID   uuid.UUID
	literal   Value
	operator  string
	function  string
	arguments []expression
}

func (source expression) clone() expression {
	result := source
	result.literal = source.literal.clone()
	result.arguments = make([]expression, len(source.arguments))
	for index := range source.arguments {
		result.arguments[index] = source.arguments[index].clone()
	}
	return result
}

type calculation struct {
	fieldID    uuid.UUID
	expression expression
}

type defaultValue struct {
	fieldID    uuid.UUID
	expression expression
}

type crossFieldRule struct {
	id        uuid.UUID
	code      string
	predicate expression
}
