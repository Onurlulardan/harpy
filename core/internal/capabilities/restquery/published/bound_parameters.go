package published

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

type boundParameter struct {
	definition ParameterDefinition
	values     []resolved.Value
}

type BoundParameters struct {
	parameters []boundParameter
	page       *PageOverride
	validated  bool
}

func (bound BoundParameters) Lookup(
	name string,
) ([]resolved.Value, bool, error) {
	if err := bound.Validate(); err != nil {
		return nil, false, err
	}
	for _, parameter := range bound.parameters {
		if parameter.definition.name == name {
			return cloneResolvedValues(parameter.values), true, nil
		}
	}
	return nil, false, nil
}

func (bound BoundParameters) Names() ([]string, error) {
	if err := bound.Validate(); err != nil {
		return nil, err
	}
	result := make([]string, len(bound.parameters))
	for index := range bound.parameters {
		result[index] = bound.parameters[index].definition.name
	}
	return result, nil
}

func (bound BoundParameters) Page() (PageOverride, bool, error) {
	if err := bound.Validate(); err != nil {
		return PageOverride{}, false, err
	}
	if bound.page == nil {
		return PageOverride{}, false, nil
	}
	return *bound.page, true, nil
}

func (BoundParameters) String() string   { return publishedRedaction }
func (BoundParameters) GoString() string { return publishedRedaction }
func (bound BoundParameters) Format(state fmt.State, _ rune) {
	writePublishedRedaction(state)
}
func (BoundParameters) MarshalJSON() ([]byte, error) { return rejectPublishedJSON() }

func newBoundParameters(
	parameters []boundParameter, page *PageOverride,
) (BoundParameters, error) {
	owned := make([]boundParameter, len(parameters))
	for index, parameter := range parameters {
		owned[index] = boundParameter{
			definition: parameter.definition.clone(),
			values:     cloneResolvedValues(parameter.values),
		}
	}
	result := BoundParameters{parameters: owned, validated: true}
	if page != nil {
		copied := *page
		result.page = &copied
	}
	if err := result.Validate(); err != nil {
		return BoundParameters{}, err
	}
	return result, nil
}
