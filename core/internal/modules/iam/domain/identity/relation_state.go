package identity

type MembershipKind string

const (
	MembershipMember MembershipKind = "member"
	MembershipOwner  MembershipKind = "owner"
)

func (kind MembershipKind) Valid() bool {
	return kind == MembershipMember || kind == MembershipOwner
}

type MembershipState string

const (
	MembershipActive     MembershipState = "active"
	MembershipSuspended  MembershipState = "suspended"
	MembershipTombstoned MembershipState = "tombstoned"
)

func (state MembershipState) Valid() bool {
	return state == MembershipActive || state == MembershipSuspended ||
		state == MembershipTombstoned
}

type AssignmentState string

const (
	AssignmentActive    AssignmentState = "active"
	AssignmentSuspended AssignmentState = "suspended"
	AssignmentRevoked   AssignmentState = "revoked"
)

func (state AssignmentState) Valid() bool {
	return state == AssignmentActive || state == AssignmentSuspended ||
		state == AssignmentRevoked
}
