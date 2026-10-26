package identity

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

type AssignmentStore[Scope any] interface {
	GrantAssignment(context.Context, Scope, identitydomain.GrantAssignmentCommand) (identitydomain.MutationResult, error)
	GetAssignment(context.Context, Scope, identitydomain.GetAssignmentQuery) (identitydomain.Assignment, error)
	ListAssignments(context.Context, Scope, identitydomain.ListAssignmentsQuery) ([]identitydomain.Assignment, error)
	UpdateAssignment(context.Context, Scope, identitydomain.UpdateAssignmentCommand) (identitydomain.MutationResult, error)
	RevokeAssignment(context.Context, Scope, identitydomain.RevokeAssignmentCommand) (identitydomain.MutationResult, error)
}
