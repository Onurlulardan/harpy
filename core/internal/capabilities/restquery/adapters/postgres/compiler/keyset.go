package compiler

import (
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func compileKeyset(
	sorts []resolved.Sort, last []cursor.LastValue, binder *binder,
) (string, error) {
	if len(sorts) == 0 || len(sorts) != len(last) || binder == nil {
		return "", ErrContract
	}
	finalNull, err := last[len(last)-1].IsNull()
	if err != nil || finalNull {
		return "", ErrContract
	}
	branches := make([]string, 0, len(sorts))
	for index := range sorts {
		currentNull, nullErr := last[index].IsNull()
		if nullErr != nil {
			return "", ErrContract
		}
		if currentNull {
			continue
		}
		branch, branchErr := keysetBranch(sorts, last, index, binder)
		if branchErr != nil {
			return "", branchErr
		}
		branches = append(branches, parenthesize(branch))
	}
	if len(branches) == 0 {
		return "", ErrContract
	}
	return strings.Join(branches, " OR "), nil
}

func keysetBranch(
	sorts []resolved.Sort,
	last []cursor.LastValue,
	index int,
	binder *binder,
) (string, error) {
	conditions := make([]string, 0, index+1)
	for prefix := 0; prefix < index; prefix++ {
		condition, err := keysetEquality(sorts[prefix], last[prefix], binder)
		if err != nil {
			return "", err
		}
		conditions = append(conditions, condition)
	}
	after, err := keysetAfter(sorts[index], last[index], binder)
	if err != nil {
		return "", err
	}
	conditions = append(conditions, after)
	return strings.Join(conditions, " AND "), nil
}

func keysetAfter(
	sortValue resolved.Sort, last cursor.LastValue, binder *binder,
) (string, error) {
	field, fieldErr := sortValue.Field()
	direction, directionErr := sortValue.Direction()
	value, present, valueErr := last.Value()
	if fieldErr != nil || directionErr != nil || valueErr != nil || !present {
		return "", ErrContract
	}
	expression, err := orderExpression(field, binder)
	if err != nil {
		return "", err
	}
	placeholder, err := bindOrderValue(field, value, binder)
	if err != nil {
		return "", err
	}
	operator := ">"
	if direction == api.DirectionDescending {
		operator = "<"
	} else if direction != api.DirectionAscending {
		return "", ErrContract
	}
	return "(" + expression + " " + operator + " " + placeholder +
		" OR " + expression + " IS NULL)", nil
}

func bindOrderValue(
	field api.FieldDefinition, value resolved.Value, binder *binder,
) (string, error) {
	_, kind, err := fieldIdentity(field)
	if err != nil {
		return "", err
	}
	if kind == api.LogicalTypeEnum {
		return bindEnumOrdinal(field, value, binder)
	}
	return bindFieldValue(field, value, binder)
}
