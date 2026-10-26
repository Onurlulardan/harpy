package resolver

import (
	"errors"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func resolveFilter(
	catalog api.FieldCatalog,
	request api.RequestFilter,
	path string,
	collector *resolutionIssues,
) (resolved.Filter, bool) {
	if predicate, found := request.Predicate(); found {
		return resolvePredicate(catalog, predicate, path, collector)
	}
	children := request.Children()
	resolvedChildren := make([]resolved.Filter, 0, len(children))
	valid := true
	for index, child := range children {
		childPath := childFilterPath(path, request.Kind(), index)
		resolvedChild, childValid := resolveFilter(catalog, child, childPath, collector)
		valid = valid && childValid
		if childValid {
			resolvedChildren = append(resolvedChildren, resolvedChild)
		}
	}
	if !valid {
		return resolved.Filter{}, false
	}
	filter, err := resolved.NewBooleanFilter(request.Kind(), resolvedChildren)
	if err != nil {
		collector.add(path, restqueryv1.IssueReferenceInvalid)
		return resolved.Filter{}, false
	}
	return filter, true
}

func resolvePredicate(
	catalog api.FieldCatalog,
	predicate api.RequestPredicate,
	path string,
	collector *resolutionIssues,
) (resolved.Filter, bool) {
	field, found, err := catalog.FieldByPublicName(predicate.Field())
	if err != nil || !found {
		collector.add(path+"/field", restqueryv1.IssueFieldUnknown)
		return resolved.Filter{}, false
	}
	capabilities, err := field.Capabilities()
	if err != nil {
		collector.add(path+"/field", restqueryv1.IssueCapabilityRequired)
		return resolved.Filter{}, false
	}
	if _, allowed, filterErr := capabilities.Filter(predicate.Operator()); filterErr != nil || !allowed {
		collector.add(path+"/op", restqueryv1.IssueOperatorNotAllowed)
		return resolved.Filter{}, false
	}
	fieldType, err := field.Type()
	if err != nil {
		collector.add(path+"/field", restqueryv1.IssueTypeMismatch)
		return resolved.Filter{}, false
	}
	values, err := ResolveLiterals(fieldType, predicate.Operator(), predicate.Values())
	if err != nil {
		collector.add(literalErrorPath(path, predicate.Operator(), err), literalIssue(err))
		return resolved.Filter{}, false
	}
	resolvedPredicate, err := resolved.NewPredicate(
		field, predicate.Operator(), values, mustRequestPath(path),
	)
	if err != nil {
		collector.add(path, restqueryv1.IssueReferenceInvalid)
		return resolved.Filter{}, false
	}
	filter, err := resolved.NewPredicateFilter(resolvedPredicate)
	if err != nil {
		collector.add(path, restqueryv1.IssueReferenceInvalid)
		return resolved.Filter{}, false
	}
	return filter, true
}

func literalIssue(err error) restqueryv1.IssueCode {
	var resolutionError LiteralResolutionError
	if errors.As(err, &resolutionError) {
		if code, codeErr := resolutionError.Code(); codeErr == nil {
			return code
		}
	}
	return restqueryv1.IssueValueNotCanonical
}

func literalErrorPath(path string, operator api.Operator, err error) string {
	code := literalIssue(err)
	if code == restqueryv1.IssueOperatorNotAllowed {
		return path + "/op"
	}
	return literalMemberPath(path, operator)
}
