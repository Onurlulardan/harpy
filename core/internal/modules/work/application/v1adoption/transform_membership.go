package v1adoption

func transformMemberships(plan *Plan, index legacyIndex) error {
	for id, membership := range index.memberships {
		source := index.source(KindMembership, id)
		role, valid := mapRole(membership.Role)
		if _, exists := index.boards[membership.BoardID]; !exists {
			appendIssue(plan, source, IssueOrphan)
			continue
		}
		if !membership.UserID.Valid() || !valid || membership.CreatedAt.IsZero() ||
			membership.UpdatedAt.IsZero() {
			appendIssue(plan, source, IssueInvalidEnum)
			continue
		}
		document := TargetMembershipDocument{BoardID: membership.BoardID.String(),
			PrincipalKind: "USER", PrincipalID: membership.UserID.String(), Role: role,
			CreatedAt: membership.CreatedAt.UTC(), UpdatedAt: membership.UpdatedAt.UTC()}
		if err := appendTarget(plan, source, TargetMembership, id, document); err != nil {
			return err
		}
	}
	return nil
}
