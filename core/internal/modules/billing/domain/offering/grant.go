package offering

import "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"

type Grant struct {
	definition entitlement.Definition
	value      entitlement.Value
}

func NewGrant(
	definition entitlement.Definition,
	value entitlement.Value,
) (Grant, error) {
	if !definition.Accepts(value) {
		return Grant{}, ErrGrant
	}
	return Grant{definition: definition, value: value}, nil
}

func (grant Grant) Definition() entitlement.Definition { return grant.definition }
func (grant Grant) Value() entitlement.Value           { return grant.value }
