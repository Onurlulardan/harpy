package invitation

import (
	"context"
	"strings"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type RoleLookup struct{ Name string }

type RoleTarget struct {
	ID    uuid.UUID
	Name  string
	Scope RoleScope
}

func (target RoleTarget) Validate() error {
	if target.ID == uuid.Nil || target.Name == "" || target.Scope.Validate() != nil {
		return ErrInvalid
	}
	return nil
}

type CredentialLookup struct {
	Selector      [16]byte
	BindingSHA256 [32]byte
}

type CredentialRecord struct {
	Invitation     Invitation
	Proof          TokenProof
	ProfileCatalog []byte
}

func (record CredentialRecord) Validate() error {
	if record.Invitation.Validate() != nil || record.Proof.Validate() != nil ||
		(record.Invitation.State != StatePending &&
			record.Invitation.State != StateAccepted) ||
		record.Proof.BindingSHA256 == [32]byte{} || len(record.ProfileCatalog) == 0 {
		return ErrInvalid
	}
	return nil
}

func (service *Service) LookupRole(
	ctx context.Context, scope platformtx.TenantScope, lookup RoleLookup,
) (RoleTarget, error) {
	lookup.Name = strings.TrimSpace(lookup.Name)
	if service == nil || service.store == nil || lookup.Name == "" ||
		len(lookup.Name) > 255 || scope.ActorID() == uuid.Nil {
		return RoleTarget{}, ErrInvalid
	}
	result, err := service.store.LookupRole(ctx, scope, lookup)
	if err != nil {
		return RoleTarget{}, err
	}
	if result.Validate() != nil {
		return RoleTarget{}, ErrUnavailable
	}
	return result, nil
}

func (service *Service) LookupCredential(
	ctx context.Context,
	scope platformtx.CredentialTenantScope,
	lookup CredentialLookup,
) (CredentialRecord, error) {
	if service == nil || service.store == nil || lookup.Selector == [16]byte{} ||
		lookup.BindingSHA256 == [32]byte{} ||
		platformtx.RequireCredentialPurpose(scope, platformtx.CredentialInvitation) != nil ||
		scope.LocatorBindingSHA256() != lookup.BindingSHA256 {
		return CredentialRecord{}, ErrInvalid
	}
	result, err := service.store.LookupCredential(ctx, scope, lookup)
	if err != nil {
		return CredentialRecord{}, err
	}
	if result.Validate() != nil || result.Proof.Selector != lookup.Selector ||
		result.Proof.BindingSHA256 != lookup.BindingSHA256 {
		return CredentialRecord{}, ErrUnavailable
	}
	return result, nil
}
