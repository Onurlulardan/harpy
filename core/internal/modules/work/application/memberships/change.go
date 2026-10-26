package memberships

import (
	"context"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type ChangeMembershipInput struct {
	BoardID, MembershipID domain.ID
	Principal             domain.Principal
	Role                  domain.BoardRole
	Active                bool
}

func (service *Service) ChangeMembership(
	ctx context.Context, scope ports.Scope, meta ports.Mutation, input ChangeMembershipInput,
) (ports.OperationReceipt, error) {
	if !input.BoardID.Valid() || !input.Principal.Valid() || !input.Role.Valid() ||
		(meta.ExpectedVersion == 0) != (!input.MembershipID.Valid()) {
		return ports.OperationReceipt{}, ports.ErrInvalid
	}
	exists, revision, err := service.principals.Exists(ctx, scope, input.Principal)
	if err != nil || !exists || revision < 1 {
		return ports.OperationReceipt{}, ports.ErrForbidden
	}
	hash := ports.Fingerprint(input.BoardID.String(), input.MembershipID.String(),
		string(input.Principal.Kind), input.Principal.ID.String(), string(input.Role),
		ports.VersionString(meta.ExpectedVersion), ports.VersionString(revision))
	return service.mutator.Mutate(ctx, scope, "work.membership.change", input.BoardID, meta, hash,
		func(ctx context.Context, stores ports.Stores) (ports.OperationReceipt, []byte, error) {
			memberships, err := stores.Boards().LoadMembershipsForUpdate(ctx, input.BoardID)
			if err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			if input.MembershipID.Valid() {
				if err = checkMembershipChange(memberships, input, meta.ExpectedVersion); err != nil {
					return ports.OperationReceipt{}, nil, err
				}
				if !input.Active {
					err = stores.Boards().DeactivateMembership(ctx, input.MembershipID, meta.ExpectedVersion)
					return membershipReceipt(input.MembershipID, meta.ExpectedVersion+1, input, err)
				}
			} else {
				input.MembershipID, err = service.ids.New(service.clock.Now().UTC())
				if err != nil {
					return ports.OperationReceipt{}, nil, ports.ErrUnavailable
				}
			}
			membership := domain.Membership{ID: input.MembershipID, BoardID: input.BoardID,
				Principal: input.Principal, Role: input.Role, Active: true,
				RowVersion: meta.ExpectedVersion + 1}
			if !membership.Valid() {
				return ports.OperationReceipt{}, nil, ports.ErrInvalid
			}
			if err = stores.Boards().UpsertMembership(ctx, membership); err != nil {
				return ports.OperationReceipt{}, nil, err
			}
			return membershipReceipt(input.MembershipID, membership.RowVersion, input, nil)
		})
}

func checkMembershipChange(items []domain.Membership, input ChangeMembershipInput, version int64) error {
	for _, item := range items {
		if item.ID != input.MembershipID {
			continue
		}
		if item.RowVersion != version {
			return ports.ErrConflict
		}
		if item.Principal.Kind == domain.PrincipalUser && item.Role == domain.RoleAdmin &&
			(!input.Active || input.Role != domain.RoleAdmin) {
			return domain.ValidateAdminRemoval(items, input.MembershipID)
		}
		return nil
	}
	return ports.ErrNotFound
}

func membershipReceipt(id domain.ID, version int64, input ChangeMembershipInput, err error) (ports.OperationReceipt, []byte, error) {
	if err != nil {
		return ports.OperationReceipt{}, nil, err
	}
	payload, _ := json.Marshal(struct {
		BoardID, MembershipID string
		Active                bool
	}{
		input.BoardID.String(), id.String(), input.Active})
	return ports.OperationReceipt{AggregateID: id, RowVersion: version}, payload, nil
}
