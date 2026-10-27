package serviceprincipal

type CapabilityDefinition struct {
	Key           CapabilityKey
	Owner         string
	ResourceKind  string
	TenantScope   bool
	ResourceScope bool
}

func (definition CapabilityDefinition) validate() error {
	if _, err := ParseCapabilityKey(string(definition.Key)); err != nil ||
		!resourceKeyPattern.MatchString(definition.Owner) ||
		!resourceKeyPattern.MatchString(definition.ResourceKind) ||
		!(definition.TenantScope || definition.ResourceScope) {
		return ErrInvalidGrant
	}
	return nil
}
