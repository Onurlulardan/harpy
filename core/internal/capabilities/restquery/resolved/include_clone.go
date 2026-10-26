package resolved

import "fmt"

func (include Include) clone() Include {
	result := include
	result.projections = append([]Projection{}, include.projections...)
	result.sorts = append([]Sort{}, include.sorts...)
	result.children = cloneIncludes(include.children)
	if include.where != nil {
		owned := include.where.clone()
		result.where = &owned
	}
	return result
}

func cloneIncludes(values []Include) []Include {
	result := make([]Include, len(values))
	for index := range values {
		result[index] = values[index].clone()
	}
	return result
}

func (Include) String() string                         { return valueRedaction }
func (Include) GoString() string                       { return valueRedaction }
func (include Include) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Include) MarshalJSON() ([]byte, error)           { return rejectDirectJSONMarshal() }
