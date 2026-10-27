package resolver

import (
	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveHaving(
	request api.RequestHaving,
	path string,
	groups []resolved.AggregateGroup,
	measures []resolved.AggregateMeasure,
	groupOrdinals map[string]int,
	measureOrdinals map[string]int,
	collector *resolutionIssues,
) (resolved.Having, bool) {
	if predicate, found := request.Predicate(); found {
		return resolveHavingPredicate(
			predicate, path, groups, measures,
			groupOrdinals, measureOrdinals, collector,
		)
	}
	children := request.Children()
	resolvedChildren := make([]resolved.Having, 0, len(children))
	valid := true
	for index, child := range children {
		childPath := childHavingPath(path, request.Kind(), index)
		resolvedChild, childValid := resolveHaving(
			child, childPath, groups, measures,
			groupOrdinals, measureOrdinals, collector,
		)
		valid = valid && childValid
		if childValid {
			resolvedChildren = append(resolvedChildren, resolvedChild)
		}
	}
	if !valid {
		return resolved.Having{}, false
	}
	having, err := resolved.NewBooleanHaving(request.Kind(), resolvedChildren)
	if err != nil {
		collector.add(path, restqueryv1.IssueReferenceInvalid)
		return resolved.Having{}, false
	}
	return having, true
}

func resolveHavingPredicate(
	predicate api.RequestHavingPredicate,
	path string,
	groups []resolved.AggregateGroup,
	measures []resolved.AggregateMeasure,
	groupOrdinals map[string]int,
	measureOrdinals map[string]int,
	collector *resolutionIssues,
) (resolved.Having, bool) {
	reference, fieldType, valid := resolveHavingReference(
		predicate, path, groups, measures, groupOrdinals, measureOrdinals, collector,
	)
	if !valid {
		return resolved.Having{}, false
	}
	values, err := ResolveLiterals(fieldType, predicate.Operator(), predicate.Values())
	if err != nil {
		collector.add(literalErrorPath(path, predicate.Operator(), err), literalIssue(err))
		return resolved.Having{}, false
	}
	resolvedPredicate, err := resolved.NewHavingPredicate(
		reference, predicate.Operator(), values,
	)
	if err != nil {
		collector.add(path, restqueryv1.IssueReferenceInvalid)
		return resolved.Having{}, false
	}
	having, err := resolved.NewPredicateHaving(resolvedPredicate)
	if err != nil {
		collector.add(path, restqueryv1.IssueReferenceInvalid)
		return resolved.Having{}, false
	}
	return having, true
}

func resolveHavingReference(
	predicate api.RequestHavingPredicate,
	path string,
	groups []resolved.AggregateGroup,
	measures []resolved.AggregateMeasure,
	groupOrdinals map[string]int,
	measureOrdinals map[string]int,
	collector *resolutionIssues,
) (resolved.AggregateReference, api.FieldType, bool) {
	if predicate.Reference() == api.AggregateReferenceGroup {
		ordinal, found := groupOrdinals[predicate.Name()]
		if !found {
			collector.add(path+"/name", restqueryv1.IssueFieldUnknown)
			return resolved.AggregateReference{}, api.FieldType{}, false
		}
		field, err := groups[ordinal].Field()
		if err != nil {
			collector.add(path+"/name", restqueryv1.IssueReferenceInvalid)
			return resolved.AggregateReference{}, api.FieldType{}, false
		}
		fieldType, err := field.Type()
		if err != nil {
			collector.add(path+"/name", restqueryv1.IssueTypeMismatch)
			return resolved.AggregateReference{}, api.FieldType{}, false
		}
		reference, err := resolved.NewAggregateReference(
			api.AggregateReferenceGroup, ordinal, mustRequestPath(path),
		)
		return reference, fieldType, err == nil
	}
	ordinal, found := measureOrdinals[predicate.Name()]
	if !found {
		collector.add(path+"/name", restqueryv1.IssueFieldUnknown)
		return resolved.AggregateReference{}, api.FieldType{}, false
	}
	fieldType, err := measures[ordinal].ResultType()
	if err != nil {
		collector.add(path+"/name", restqueryv1.IssueTypeMismatch)
		return resolved.AggregateReference{}, api.FieldType{}, false
	}
	reference, err := resolved.NewAggregateReference(
		api.AggregateReferenceMeasure, ordinal, mustRequestPath(path),
	)
	return reference, fieldType, err == nil
}
