package systemresource

import identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"

type Assignment = identitydomain.Assignment
type Membership = identitydomain.Membership
type AssignmentState = identitydomain.AssignmentState
type MembershipState = identitydomain.MembershipState
type ListAssignmentsQuery = identitydomain.ListAssignmentsQuery
type ListMembershipsQuery = identitydomain.ListMembershipsQuery
type RevokeAssignmentCommand = identitydomain.RevokeAssignmentCommand
type DeleteMembershipCommand = identitydomain.DeleteMembershipCommand
type UserStateCommand = identitydomain.UserStateCommand
type MutationResult = identitydomain.MutationResult

const (
	AssignmentActive    = identitydomain.AssignmentActive
	AssignmentSuspended = identitydomain.AssignmentSuspended
	AssignmentRevoked   = identitydomain.AssignmentRevoked

	MembershipActive     = identitydomain.MembershipActive
	MembershipSuspended  = identitydomain.MembershipSuspended
	MembershipTombstoned = identitydomain.MembershipTombstoned
)
