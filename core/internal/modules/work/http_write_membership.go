package work

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/memberships"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (runtime *PlatformHTTPRuntime) writeMembership(ctx context.Context, scope ports.Scope,
	input HTTPWriteInput, wire writeWire,
) (ports.OperationReceipt, bool, error) {
	switch input.Operation {
	case "changeWorkMembership", "createLegacyBoardMember",
		"updateLegacyBoardMember", "deleteLegacyBoardMember":
	default:
		return ports.OperationReceipt{}, false, nil
	}
	boardID, err := routeOrBodyID(input.IDs, "board_id", wire.BoardID)
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	if input.Operation == "updateLegacyBoardMember" ||
		input.Operation == "deleteLegacyBoardMember" {
		return runtime.updateLegacyMembership(ctx, scope, input, wire, boardID)
	}
	membershipID, _ := optionalValueID(wire.MembershipID)
	principalID := wire.PrincipalID
	if principalID == "" {
		principalID = wire.UserID
	}
	principal, err := principalValue(wire.PrincipalKind, principalID)
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	active := true
	if wire.Active != nil {
		active = *wire.Active
	}
	receipt, err := runtime.services.memberships.ChangeMembership(ctx, scope,
		mutation(input, wire.ExpectedVersion), memberships.ChangeMembershipInput{
			BoardID: boardID, MembershipID: membershipID, Principal: principal,
			Role: ports.BoardRole(wire.Role), Active: active})
	return receipt, true, err
}

func (runtime *PlatformHTTPRuntime) updateLegacyMembership(ctx context.Context,
	scope ports.Scope, input HTTPWriteInput, wire writeWire, boardID ports.ID,
) (ports.OperationReceipt, bool, error) {
	membershipID, err := routeID(input.IDs, "member_id")
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	stores, err := runtime.stores(ctx, scope)
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	items, err := stores.Boards().LoadMembershipsForUpdate(ctx, boardID)
	if err != nil {
		return ports.OperationReceipt{}, true, err
	}
	for _, item := range items {
		if item.ID != membershipID {
			continue
		}
		role, active := ports.BoardRole(wire.Role), true
		if !role.Valid() {
			role = item.Role
		}
		if input.Operation == "deleteLegacyBoardMember" {
			active = false
		}
		receipt, writeErr := runtime.services.memberships.ChangeMembership(ctx, scope,
			mutation(input, item.RowVersion), memberships.ChangeMembershipInput{
				BoardID: boardID, MembershipID: item.ID, Principal: item.Principal,
				Role: role, Active: active})
		return receipt, true, writeErr
	}
	return ports.OperationReceipt{}, true, ports.ErrNotFound
}

func optionalValueID(value string) (ports.ID, error) {
	if value == "" {
		return ports.ID{}, nil
	}
	return parseID(value)
}

func principalValue(kind, id string) (ports.Principal, error) {
	if kind == "" {
		kind = string(ports.PrincipalUser)
	}
	parsed, err := parseID(id)
	value := ports.Principal{Kind: ports.PrincipalKind(kind), ID: parsed}
	if err != nil || !value.Valid() {
		return ports.Principal{}, ports.ErrInvalid
	}
	return value, nil
}
