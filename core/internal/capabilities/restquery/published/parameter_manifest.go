package published

import "fmt"

const maximumPublishedParameters = 100

type ParameterManifest struct {
	parameters      []ParameterDefinition
	maximumPageSize int
	validated       bool
}

func NewParameterManifest(
	inputs []ParameterInput, maximumPageSize int,
) (ParameterManifest, error) {
	if len(inputs) > maximumPublishedParameters {
		return ParameterManifest{}, fmt.Errorf("published parameter manifest is invalid")
	}
	parameters := make([]ParameterDefinition, 0, len(inputs))
	seen := make(map[string]struct{}, len(inputs))
	for _, input := range inputs {
		if _, duplicate := seen[input.Name]; duplicate {
			return ParameterManifest{}, fmt.Errorf("published parameter manifest is invalid")
		}
		fieldType, err := cloneParameterFieldType(input.Type)
		if err != nil {
			return ParameterManifest{}, fmt.Errorf("published parameter manifest is invalid")
		}
		definition := ParameterDefinition{
			name: input.Name, fieldType: fieldType, arity: input.Arity,
			required: input.Required, defaults: cloneResolvedValues(input.Default),
			validated: true,
		}
		if err := definition.Validate(); err != nil {
			return ParameterManifest{}, err
		}
		seen[input.Name] = struct{}{}
		parameters = append(parameters, definition)
	}
	manifest := ParameterManifest{
		parameters: parameters, maximumPageSize: maximumPageSize, validated: true,
	}
	if err := manifest.Validate(); err != nil {
		return ParameterManifest{}, err
	}
	return manifest, nil
}

func (manifest ParameterManifest) Parameters() ([]ParameterDefinition, error) {
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	result := make([]ParameterDefinition, len(manifest.parameters))
	for index := range manifest.parameters {
		result[index] = manifest.parameters[index].clone()
	}
	return result, nil
}

func (manifest ParameterManifest) Lookup(
	name string,
) (ParameterDefinition, bool, error) {
	if err := manifest.Validate(); err != nil {
		return ParameterDefinition{}, false, err
	}
	for _, definition := range manifest.parameters {
		if definition.name == name {
			return definition.clone(), true, nil
		}
	}
	return ParameterDefinition{}, false, nil
}

func (manifest ParameterManifest) MaximumPageSize() (int, error) {
	if err := manifest.Validate(); err != nil {
		return 0, err
	}
	return manifest.maximumPageSize, nil
}

func (manifest ParameterManifest) String() string   { return publishedRedaction }
func (manifest ParameterManifest) GoString() string { return publishedRedaction }
func (manifest ParameterManifest) Format(state fmt.State, _ rune) {
	writePublishedRedaction(state)
}
func (ParameterManifest) MarshalJSON() ([]byte, error) { return rejectPublishedJSON() }
