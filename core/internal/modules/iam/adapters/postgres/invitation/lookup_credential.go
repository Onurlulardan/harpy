package postgres

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
)

func (store *invitationStore) LookupCredential(
	ctx context.Context,
	scope credentialScope,
	lookup invitation.CredentialLookup,
) (invitation.CredentialRecord, error) {
	if !invitationCredentialScopeUsable(ctx, scope) ||
		scope.LocatorBindingSHA256() != lookup.BindingSHA256 {
		return invitation.CredentialRecord{}, invitation.ErrInvalid
	}
	bound, err := store.bindCredential(scope)
	if err != nil {
		return invitation.CredentialRecord{}, err
	}
	var result invitation.CredentialRecord
	var selector, secretHMAC, binding []byte
	row := bound.database.QueryRow(
		ctx, invitationCredentialLookupSQL, lookup.Selector[:], lookup.BindingSHA256[:],
	)
	err = scanInvitationCredential(
		row, &result, &selector, &secretHMAC, &binding,
	)
	defer clear(selector)
	defer clear(secretHMAC)
	defer clear(binding)
	if err != nil {
		return invitation.CredentialRecord{}, err
	}
	if len(selector) != 16 || len(secretHMAC) != 32 || len(binding) != 32 {
		return invitation.CredentialRecord{}, invitation.ErrUnavailable
	}
	copy(result.Proof.Selector[:], selector)
	copy(result.Proof.SecretHMAC[:], secretHMAC)
	copy(result.Proof.BindingSHA256[:], binding)
	if result.Validate() != nil {
		return invitation.CredentialRecord{}, invitation.ErrUnavailable
	}
	return result, nil
}
