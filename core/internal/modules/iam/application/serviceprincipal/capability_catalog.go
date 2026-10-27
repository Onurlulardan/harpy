package serviceprincipal

type CapabilityCatalog struct {
	definitions map[CapabilityKey]CapabilityDefinition
	valid       bool
}

func NewCapabilityCatalog(
	definitions []CapabilityDefinition,
) (CapabilityCatalog, error) {
	if len(definitions) == 0 || len(definitions) > 1000 {
		return CapabilityCatalog{}, ErrInvalidGrant
	}
	catalog := CapabilityCatalog{
		definitions: make(map[CapabilityKey]CapabilityDefinition, len(definitions)),
		valid:       true,
	}
	for _, definition := range definitions {
		if definition.validate() != nil {
			return CapabilityCatalog{}, ErrInvalidGrant
		}
		if _, duplicate := catalog.definitions[definition.Key]; duplicate {
			return CapabilityCatalog{}, ErrInvalidGrant
		}
		catalog.definitions[definition.Key] = definition
	}
	return catalog, nil
}

func (catalog CapabilityCatalog) Allows(
	capability CapabilityKey,
	scope ResourceScope,
) bool {
	if !catalog.valid || scope.Validate() != nil {
		return false
	}
	definition, exists := catalog.definitions[capability]
	if !exists || definition.Owner != scope.Owner ||
		definition.ResourceKind != scope.ResourceKind {
		return false
	}
	return scope.Kind == ScopeTenant && definition.TenantScope ||
		scope.Kind == ScopeResource && definition.ResourceScope
}
