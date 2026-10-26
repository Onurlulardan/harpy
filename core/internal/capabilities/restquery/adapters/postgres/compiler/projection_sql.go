package compiler

import "strings"

func appendProjectionList(builder *strings.Builder, columns []recordColumn) error {
	if builder == nil || len(columns) == 0 {
		return ErrContract
	}
	for index, column := range columns {
		if index > 0 {
			builder.WriteString(", ")
		}
		expression, err := baseExpression(column.field)
		if err != nil {
			return ErrContract
		}
		builder.WriteString(expression)
	}
	return nil
}
