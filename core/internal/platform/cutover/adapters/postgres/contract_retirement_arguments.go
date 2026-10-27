package postgres

import "github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"

func contractRetirementIdentityArguments(
	key contractretirement.Key, identity contractretirement.Identity,
) (string, string, string, error) {
	keyValue, err := key.DatabaseValue()
	if err != nil {
		return "", "", "", err
	}
	owner, domain, err := identity.DatabaseValues()
	if err != nil {
		return "", "", "", err
	}
	return keyValue, owner, domain, nil
}
