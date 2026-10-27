package published

import "fmt"

func (bound BoundParameters) Validate() error {
	if !bound.validated || bound.parameters == nil {
		return fmt.Errorf("published REST Query bound parameters are invalid")
	}
	seen := make(map[string]struct{}, len(bound.parameters))
	for _, parameter := range bound.parameters {
		if parameter.definition.Validate() != nil ||
			!validateResolvedParameterValues(
				parameter.definition.fieldType,
				parameter.definition.arity,
				parameter.values,
			) {
			return fmt.Errorf("published REST Query bound parameters are invalid")
		}
		if _, duplicate := seen[parameter.definition.name]; duplicate {
			return fmt.Errorf("published REST Query bound parameters are invalid")
		}
		seen[parameter.definition.name] = struct{}{}
	}
	if bound.page != nil && bound.page.Validate() != nil {
		return fmt.Errorf("published REST Query bound parameters are invalid")
	}
	return nil
}
