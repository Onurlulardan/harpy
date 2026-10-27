package published

import (
	"fmt"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func BindGET(
	manifest ParameterManifest,
	provided map[string][]api.Literal,
	page *PageOverride,
) (BoundParameters, error) {
	return bindParameters(manifest, provided, page, "/param")
}

func BindPOST(
	manifest ParameterManifest,
	provided map[string][]api.Literal,
	page *PageOverride,
) (BoundParameters, error) {
	return bindParameters(manifest, provided, page, "/params")
}

func bindParameters(
	manifest ParameterManifest,
	provided map[string][]api.Literal,
	page *PageOverride,
	root string,
) (BoundParameters, error) {
	if manifest.Validate() != nil || page != nil && page.Validate() != nil {
		return BoundParameters{}, fmt.Errorf("published parameter binding is invalid")
	}
	collector := bindingIssueCollector{}
	for name := range provided {
		if _, found, _ := manifest.Lookup(name); !found {
			collector.add(
				parameterPointer(root, name), restqueryv1.IssueParameterUnknown,
			)
		}
	}
	bound := make([]boundParameter, 0, len(manifest.parameters))
	for _, definition := range manifest.parameters {
		literals, present := provided[definition.name]
		if !present {
			bindMissingParameter(definition, root, &bound, &collector)
			continue
		}
		path := parameterPointer(root, definition.name)
		if !validBindingArity(definition.arity, len(literals)) {
			collector.add(path, restqueryv1.IssueParameterArityMismatch)
			continue
		}
		values, valid := resolveParameterBinding(
			definition, literals, path, &collector,
		)
		if valid {
			bound = append(bound, boundParameter{
				definition: definition.clone(), values: values,
			})
		}
	}
	validatePageMaximum(manifest, page, &collector)
	if collector.hasIssues() {
		return BoundParameters{}, collector.queryError()
	}
	result, err := newBoundParameters(bound, page)
	if err != nil {
		return BoundParameters{}, fmt.Errorf("published parameter binding is invalid")
	}
	return result, nil
}

func bindMissingParameter(
	definition ParameterDefinition,
	root string,
	bound *[]boundParameter,
	collector *bindingIssueCollector,
) {
	if definition.required {
		collector.add(
			parameterPointer(root, definition.name),
			restqueryv1.IssueParameterRequired,
		)
		return
	}
	if definition.defaults != nil {
		*bound = append(*bound, boundParameter{
			definition: definition.clone(),
			values:     cloneResolvedValues(definition.defaults),
		})
	}
}

func validBindingArity(arity ParameterArity, count int) bool {
	return arity == ParameterAritySingle && count == 1 ||
		arity == ParameterArityList && count >= 1 && count <= 100
}

func validatePageMaximum(
	manifest ParameterManifest,
	page *PageOverride,
	collector *bindingIssueCollector,
) {
	if page == nil {
		return
	}
	size, present, err := page.Size()
	if err != nil {
		collector.add("/page", restqueryv1.IssueValueNotAllowed)
		return
	}
	if present && size > manifest.maximumPageSize {
		collector.add("/page/size", restqueryv1.IssuePageSizeExceeded)
	}
}
