package systemtransfer

import (
	"fmt"

	tabularmapping "github.com/Onurlulardan/harpy/core/internal/capabilities/tabularmapping/api"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

type MappingPolicy struct {
	resource publicresource.Resource
	allowed  map[string]struct{}
	required []string
}

func NewMappingPolicy(catalog publicresource.FieldCatalog) (MappingPolicy, error) {
	if catalog.Validate() != nil {
		return MappingPolicy{}, ErrInvalidMapping
	}
	resource, err := catalog.Resource()
	if err != nil {
		return MappingPolicy{}, ErrInvalidMapping
	}
	descriptors, err := catalog.Descriptors()
	if err != nil {
		return MappingPolicy{}, ErrInvalidMapping
	}
	allowed := make([]string, 0, len(descriptors))
	for _, descriptor := range descriptors {
		id, idErr := descriptor.ID()
		editable, editErr := catalog.Editable(id)
		if idErr != nil || editErr != nil {
			return MappingPolicy{}, ErrInvalidMapping
		}
		if editable && !protectedCoreTarget(resource, id.String()) {
			allowed = append(allowed, id.String())
		}
	}
	return newMappingPolicy(resource, allowed)
}

func newMappingPolicy(resource publicresource.Resource,
	allowedTargets []string) (MappingPolicy, error) {
	if !resource.Valid() {
		return MappingPolicy{}, ErrInvalidMapping
	}
	policy := MappingPolicy{resource: resource, allowed: make(map[string]struct{})}
	for _, target := range allowedTargets {
		if _, err := publicresource.ParseID(target); err != nil ||
			protectedCoreTarget(resource, target) {
			return MappingPolicy{}, ErrForbiddenTarget
		}
		policy.allowed[target] = struct{}{}
	}
	manifest, err := ManifestFor(resource)
	if err != nil {
		return MappingPolicy{}, err
	}
	for _, identity := range manifest.Fields {
		policy.allowed[identity] = struct{}{}
	}
	policy.required = createRequiredTargets(resource)
	for _, required := range policy.required {
		policy.allowed[required] = struct{}{}
	}
	return policy, nil
}

func (policy MappingPolicy) Compile(document []byte) (tabularmapping.Program, error) {
	if !policy.resource.Valid() || len(policy.allowed) == 0 {
		return tabularmapping.Program{}, ErrInvalidMapping
	}
	program, err := tabularmapping.Compile(document)
	if err != nil || program.Contract() != tabularmapping.ContractV2 {
		return tabularmapping.Program{}, ErrInvalidMapping
	}
	seen := make(map[string]struct{}, len(program.Targets()))
	for _, target := range program.Targets() {
		if _, parseErr := publicresource.ParseID(target); parseErr != nil {
			return tabularmapping.Program{}, ErrForbiddenTarget
		}
		if _, allowed := policy.allowed[target]; !allowed {
			return tabularmapping.Program{}, ErrForbiddenTarget
		}
		seen[target] = struct{}{}
	}
	for _, required := range policy.required {
		if _, found := seen[required]; !found {
			return tabularmapping.Program{}, fmt.Errorf("%w: %s", ErrInviteRequired, required)
		}
	}
	return program, nil
}
