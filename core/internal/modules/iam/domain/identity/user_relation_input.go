package identity

const (
	UserRelationCreateCommandKey = "iam.identity.user.create"
	UserRelationUpdateCommandKey = "iam.identity.user.update"
)

type ReplaceUserRelationsCommand struct {
	UserID           IdentityID
	ExpectedRevision int64
	CommandKey       string
	PatchProfile     bool
	ProfilePatch     []byte
	Organizations    OrganizationSetPatch
	Roles            RoleSetPatch
	Evidence         MutationEvidence
	preparedProfile  ProfileDocument
}

func (command ReplaceUserRelationsCommand) Validate() error {
	if command.UserID == (IdentityID{}) || command.ExpectedRevision < 1 ||
		command.Evidence.Validate() != nil ||
		command.Organizations.validate(command.UserID) != nil ||
		command.Roles.validate(command.UserID) != nil ||
		!validUserRelationCommandKey(command.CommandKey, command.ExpectedRevision) ||
		!command.PatchProfile && !command.Organizations.Replace() && !command.Roles.Replace() {
		return ErrInvalidCommand
	}
	if command.PatchProfile {
		if !ValidProfileInput(command.ProfilePatch) {
			return ErrInvalidCommand
		}
	} else if len(command.ProfilePatch) != 0 || !emptyProfile(command.preparedProfile) {
		return ErrInvalidCommand
	}
	if (command.Organizations.Replace() || command.Roles.Replace()) &&
		command.Evidence.ActorID == command.UserID {
		return ErrInvalidCommand
	}
	return nil
}

func (command ReplaceUserRelationsCommand) WithProfile(
	profile ProfileDocument,
) ReplaceUserRelationsCommand {
	command.preparedProfile = profile
	return command
}

func (command ReplaceUserRelationsCommand) Profile() ProfileDocument {
	return command.preparedProfile
}

func validUserRelationCommandKey(commandKey string, expectedRevision int64) bool {
	if commandKey == UserRelationCreateCommandKey {
		return expectedRevision == 1
	}
	return commandKey == UserRelationUpdateCommandKey
}

func emptyProfile(profile ProfileDocument) bool {
	return profile.Resource() == "" && profile.SchemaID() == (IdentityID{}) &&
		len(profile.Values()) == 0 && profile.SHA256() == [32]byte{}
}
