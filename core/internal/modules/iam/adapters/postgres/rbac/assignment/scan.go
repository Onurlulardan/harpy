package assignmentpostgres

import (
	"encoding/json"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/google/uuid"
)

func scanMutation(row scanner) (rbacdomain.MutationResult, error) {
	var result rbacdomain.MutationResult
	var id uuid.UUID
	if row == nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	if err := row.Scan(
		&id, &result.Revision, &result.AuthorizationClock, &result.Replayed,
	); err != nil {
		return rbacdomain.MutationResult{}, mapError(err, rbacdomain.ErrInvalidMutation)
	}
	result.ID = rbacdomain.ID(id)
	if result.Validate(result.ID) != nil {
		return rbacdomain.MutationResult{}, rbacdomain.ErrStorage
	}
	return result, nil
}

func scanAssignment(row scanner) (rbacdomain.Assignment, error) {
	var result rbacdomain.Assignment
	var id, resourceID, targetID uuid.UUID
	var target string
	var itemsJSON []byte
	if row == nil {
		return rbacdomain.Assignment{}, rbacdomain.ErrStorage
	}
	if err := row.Scan(
		&id, &resourceID, &target, &targetID, &result.State,
		&result.Revision, &result.CreatedAt, &result.UpdatedAt, &itemsJSON,
	); err != nil {
		return rbacdomain.Assignment{}, mapError(err, rbacdomain.ErrInvalidQuery)
	}
	var members []memberDocument
	if json.Unmarshal(itemsJSON, &members) != nil || len(members) == 0 {
		return rbacdomain.Assignment{}, rbacdomain.ErrStorage
	}
	result.ID, result.ResourceID = rbacdomain.ID(id), rbacdomain.ID(resourceID)
	result.Target, result.TargetID = rbacdomain.AssignmentTarget(target), rbacdomain.ID(targetID)
	result.Items = make([]rbacdomain.AssignmentItem, len(members))
	for index, member := range members {
		result.Items[index] = rbacdomain.AssignmentItem{
			Ordinal: member.Ordinal, ActionID: rbacdomain.ID(member.ActionID),
			PermissionID: rbacdomain.ID(member.PermissionID),
			BindingID:    rbacdomain.ID(member.BindingID),
		}
	}
	if !assignmentValid(result) {
		return rbacdomain.Assignment{}, rbacdomain.ErrStorage
	}
	return result, nil
}

func assignmentValid(value rbacdomain.Assignment) bool {
	return value.ID.Valid() && value.ResourceID.Valid() && value.Target.Valid() &&
		value.TargetID.Valid() && value.Revision > 0 &&
		(value.State == "active" || value.State == "retired") &&
		!value.CreatedAt.IsZero() && !value.UpdatedAt.IsZero() &&
		!value.UpdatedAt.Before(value.CreatedAt) && len(value.Items) > 0
}

func scanEffective(row scanner) (rbacdomain.EffectivePermission, error) {
	var result rbacdomain.EffectivePermission
	var permissionID, resourceID, sourceID uuid.UUID
	var actionID *uuid.UUID
	if row == nil {
		return rbacdomain.EffectivePermission{}, rbacdomain.ErrStorage
	}
	if err := row.Scan(
		&permissionID, &result.MatchKind, &resourceID, &result.ResourceKey,
		&actionID, &result.ActionKey, &result.Source, &sourceID,
	); err != nil {
		return rbacdomain.EffectivePermission{}, mapError(err, rbacdomain.ErrInvalidQuery)
	}
	result.PermissionID = rbacdomain.ID(permissionID)
	result.ResourceID = rbacdomain.ID(resourceID)
	result.SourceID = rbacdomain.ID(sourceID)
	if actionID != nil {
		result.ActionID = rbacdomain.ID(*actionID)
	}
	if !result.PermissionID.Valid() || !result.Source.Valid() ||
		!result.SourceID.Valid() || result.ResourceKey == "" ||
		(result.MatchKind == "exact" && (!result.ResourceID.Valid() ||
			!result.ActionID.Valid() || result.ActionKey == "")) ||
		(result.MatchKind != "exact" && result.MatchKind != "resource_all_actions" &&
			result.MatchKind != "tenant_all") {
		return rbacdomain.EffectivePermission{}, rbacdomain.ErrStorage
	}
	return result, nil
}
