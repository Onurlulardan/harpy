package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	roleapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/role"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
)

func (port *identitySystemResourcePort) updateSystemRole(
	command identitySystemResourceCommand,
	update identitySystemResourceUpdate,
	values identitySystemResourceValues,
) (publicresource.MutationResult, error) {
	current, err := port.application.identities.roles.GetRole(
		command.ctx, command.scope, roleapp.GetRoleQuery{
			ID:       identityfacade.IdentityID(update.id),
			Evidence: identityQueryEvidence(command.evidence, command.readPermissionID),
		},
	)
	if err != nil {
		return publicresource.MutationResult{}, identitySystemResourceError(err)
	}
	if current.Revision != update.expectedRevision {
		return publicresource.MutationResult{}, publicresource.ErrConflict
	}
	name, description, defaultValue := current.Name, current.Description, current.Default
	if value, found, fieldErr := values.text(publicresource.RoleNameFieldID, false); fieldErr != nil {
		return publicresource.MutationResult{}, fieldErr
	} else if found {
		name = value
	}
	if value, found, fieldErr := values.text(publicresource.RoleDescriptionFieldID, false); fieldErr != nil {
		return publicresource.MutationResult{}, fieldErr
	} else if found {
		description = value
	}
	if value, found, fieldErr := values.boolean(publicresource.RoleDefaultFieldID); fieldErr != nil {
		return publicresource.MutationResult{}, fieldErr
	} else if found {
		defaultValue = value
	}
	if err = identitySystemRoleScopeUnchanged(values, current); err != nil {
		return publicresource.MutationResult{}, err
	}
	profile, err := port.updateNameProfile(
		command, identityfacade.ProfileRole,
		current.Profile.SchemaID(), values, name,
	)
	if err != nil {
		return publicresource.MutationResult{}, err
	}
	return port.updateSystemRoleCommand(
		command, update, name, description, defaultValue, profile,
	)
}
