package profile

import (
	"context"
	"encoding/json"

	profilecommand "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func (store *commandStore) GetSelf(
	ctx context.Context,
	scope tenantScope,
	evidence profilecommand.SelfEvidence,
) (profilecommand.SelfRecord, error) {
	if evidence.Validate() != nil || !readScopeUsable(ctx, scope) ||
		scope.ActorID() != evidence.ActorID {
		return profilecommand.SelfRecord{}, profilecommand.ErrInvalidCommand
	}
	bound, err := store.bind(scope)
	if err != nil {
		return profilecommand.SelfRecord{}, err
	}
	var record profilecommand.SelfRecord
	var roles, organizations []byte
	err = bound.database.QueryRow(
		ctx, getSelfSQL, evidence.ActorID, evidence.ExpectedTokenVersion,
		evidence.ExpectedAuthorizationVersion,
	).Scan(
		&record.ID, &record.Email, &record.Status, &record.Revision,
		&record.CreatedAt, &record.UpdatedAt, &roles, &organizations,
	)
	if err != nil {
		return profilecommand.SelfRecord{}, mapCommandError(err)
	}
	if json.Unmarshal(roles, &record.Roles) != nil ||
		json.Unmarshal(organizations, &record.Organizations) != nil ||
		record.Validate(evidence.ActorID) != nil {
		return profilecommand.SelfRecord{}, profilecommand.ErrEvidenceMissing
	}
	return record, nil
}
