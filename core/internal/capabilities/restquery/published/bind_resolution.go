package published

import (
	"errors"
	"strconv"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolver"
)

func resolveParameterBinding(
	definition ParameterDefinition,
	literals []api.Literal,
	path string,
	collector *bindingIssueCollector,
) ([]resolved.Value, bool) {
	operator := api.OperatorEq
	if definition.arity == ParameterArityList {
		operator = api.OperatorIn
	}
	values, err := resolver.ResolveLiterals(definition.fieldType, operator, literals)
	if err == nil {
		return values, true
	}
	if definition.arity == ParameterArityList {
		invalidItem := false
		for index, literal := range literals {
			_, itemErr := resolver.ResolveLiterals(
				definition.fieldType, api.OperatorEq, []api.Literal{literal},
			)
			if itemErr != nil {
				collector.add(
					path+"/"+strconv.Itoa(index), resolutionIssue(itemErr),
				)
				invalidItem = true
			}
		}
		if invalidItem {
			return nil, false
		}
	}
	collector.add(path, resolutionIssue(err))
	return nil, false
}

func resolutionIssue(err error) restqueryv1.IssueCode {
	var resolutionError resolver.LiteralResolutionError
	if errors.As(err, &resolutionError) {
		code, codeErr := resolutionError.Code()
		if codeErr == nil && code.Valid() {
			return code
		}
	}
	return restqueryv1.IssueValueNotCanonical
}
