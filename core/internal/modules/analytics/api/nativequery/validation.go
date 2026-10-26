package nativequery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
)

var (
	identifierPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
	slugPattern       = regexp.MustCompile(`^[a-z][a-z0-9-]{0,99}$`)
)

func (plan Plan) Validate() error {
	if plan.Contract != ContractV1 || len(plan.Sources) < 2 ||
		len(plan.Sources) > 4 || len(plan.Joins) != len(plan.Sources)-1 ||
		len(plan.Fields) == 0 || len(plan.Fields) > 64 ||
		len(plan.Order) > 8 || plan.Limit < 1 || plan.Limit > 500 {
		return fmt.Errorf("Analytics native query shape is invalid")
	}
	sources, sourceSlugs := map[string]int{}, map[string]bool{}
	fields := map[string]Field{}
	for index, source := range plan.Sources {
		if !identifierPattern.MatchString(source.Alias) ||
			!slugPattern.MatchString(source.FormSlug) ||
			!jsonObject(source.Query) || sources[source.Alias] != 0 ||
			sourceSlugs[source.FormSlug] {
			return fmt.Errorf("Analytics native query source is invalid")
		}
		sources[source.Alias] = index + 1
		sourceSlugs[source.FormSlug] = true
	}
	for _, field := range plan.Fields {
		if !validField(field, sources) || fields[field.Alias].Alias != "" {
			return fmt.Errorf("Analytics native query field is invalid")
		}
		fields[field.Alias] = field
	}
	targets := map[string]bool{}
	for _, join := range plan.Joins {
		if join.Kind != "left" || !identifierPattern.MatchString(join.SourceField) ||
			!identifierPattern.MatchString(join.TargetField) ||
			sources[join.Source] == 0 || sources[join.Target] == 0 ||
			sources[join.Source] >= sources[join.Target] || targets[join.Target] {
			return fmt.Errorf("Analytics native query join is invalid")
		}
		targets[join.Target] = true
	}
	measures := map[string]bool{}
	if plan.Aggregation != nil {
		if err := validateAggregation(*plan.Aggregation, sources, fields, measures); err != nil {
			return err
		}
	}
	for _, order := range plan.Order {
		fieldOrder := order.Source != "" && order.Field != "" && order.Measure == "" &&
			hasSourceField(plan.Fields, order.Source, order.Field)
		measureOrder := order.Source == "" && order.Field == "" && measures[order.Measure]
		if (order.Direction != "asc" && order.Direction != "desc") ||
			(!fieldOrder && !measureOrder) {
			return fmt.Errorf("Analytics native query order is invalid")
		}
	}
	return nil
}

func validateAggregation(
	value Aggregation, sources map[string]int, fields map[string]Field,
	measures map[string]bool,
) error {
	if len(value.Groups) == 0 || len(value.Groups) > 8 ||
		len(value.Measures) == 0 || len(value.Measures) > 8 {
		return fmt.Errorf("Analytics native query aggregation is invalid")
	}
	groupAliases := map[string]bool{}
	for _, group := range value.Groups {
		if !validField(group, sources) || fields[group.Alias] != group ||
			groupAliases[group.Alias] {
			return fmt.Errorf("Analytics native query group is invalid")
		}
		groupAliases[group.Alias] = true
	}
	for _, measure := range value.Measures {
		allowed := measure.Function == "count" || measure.Function == "sum" ||
			measure.Function == "avg" || measure.Function == "min" ||
			measure.Function == "max"
		if !allowed || !identifierPattern.MatchString(measure.Alias) ||
			groupAliases[measure.Alias] || measures[measure.Alias] ||
			(measure.Function != "count" &&
				(sources[measure.Source] == 0 ||
					!identifierPattern.MatchString(measure.Field))) {
			return fmt.Errorf("Analytics native query measure is invalid")
		}
		measures[measure.Alias] = true
	}
	return nil
}

func validField(value Field, sources map[string]int) bool {
	return identifierPattern.MatchString(value.Alias) &&
		sources[value.Source] > 0 && identifierPattern.MatchString(value.Field)
}

func hasSourceField(fields []Field, source, field string) bool {
	for _, value := range fields {
		if value.Source == source && value.Field == field {
			return true
		}
	}
	return false
}

func jsonObject(value []byte) bool {
	value = bytes.TrimSpace(value)
	return len(value) >= 2 && len(value) <= 64<<10 &&
		value[0] == '{' && value[len(value)-1] == '}' && json.Valid(value)
}
