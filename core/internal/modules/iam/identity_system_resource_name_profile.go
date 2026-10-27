package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
)

const identityDisplayNameProfileKey = "display_name"

func (port *identitySystemResourcePort) createNameProfile(
	command identitySystemResourceCommand,
	resource identityfacade.ProfileResource,
	values identitySystemResourceValues,
	name string,
) ([]byte, error) {
	aliases, err := port.application.identities.profiles.FieldIDs(
		command.ctx, command.scope, resource, []string{identityDisplayNameProfileKey},
	)
	if err != nil {
		return nil, identitySystemResourceError(err)
	}
	fieldID, found := aliases[identityDisplayNameProfileKey]
	if !found || len(aliases) != 1 {
		return nil, publicresource.ErrUnavailable
	}
	return values.profileWithDisplayName(fieldID, name)
}

func (port *identitySystemResourcePort) updateNameProfile(
	command identitySystemResourceCommand,
	resource identityfacade.ProfileResource,
	schemaID identityfacade.IdentityID,
	values identitySystemResourceValues,
	name string,
) ([]byte, error) {
	aliases, err := port.application.identities.profiles.FieldIDsForSchema(
		command.ctx, command.scope, resource, schemaID,
		[]string{identityDisplayNameProfileKey},
	)
	if err != nil {
		return nil, identitySystemResourceError(err)
	}
	fieldID, found := aliases[identityDisplayNameProfileKey]
	if !found || len(aliases) != 1 {
		return nil, publicresource.ErrUnavailable
	}
	return values.profileWithDisplayName(fieldID, name)
}
