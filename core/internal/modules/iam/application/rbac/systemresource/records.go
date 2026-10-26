package systemresource

import rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"

type Action = rbacdomain.Action
type ActionInput = rbacdomain.ActionInput
type Assignment = rbacdomain.Assignment
type AssignmentGetQuery = rbacdomain.AssignmentGetQuery
type AssignmentItem = rbacdomain.AssignmentItem
type AssignmentListQuery = rbacdomain.AssignmentListQuery
type AssignmentMutation = rbacdomain.AssignmentMutation
type AssignmentTarget = rbacdomain.AssignmentTarget
type EffectivePermission = rbacdomain.EffectivePermission
type EffectivePermissionQuery = rbacdomain.EffectivePermissionQuery
type GetQuery = rbacdomain.GetQuery
type ID = rbacdomain.ID
type ListQuery = rbacdomain.ListQuery
type MutationEvidence = rbacdomain.MutationEvidence
type QueryEvidence = rbacdomain.QueryEvidence
type Resource = rbacdomain.Resource
type ResourceInput = rbacdomain.ResourceInput
type StateCommand = rbacdomain.StateCommand
type StepUpEvidence = rbacdomain.StepUpEvidence
type UpdateCatalog = rbacdomain.UpdateCatalog

const (
	AssignmentUser         = rbacdomain.AssignmentUser
	AssignmentRole         = rbacdomain.AssignmentRole
	AssignmentOrganization = rbacdomain.AssignmentOrganization
)
