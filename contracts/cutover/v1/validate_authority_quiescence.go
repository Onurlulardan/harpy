package cutoverv1

const authorityQuiescedPrincipal = "harpy_migrator"

func validateAuthorityQuiescencePolicy(
	policy AuthorityQuiescencePolicy, objects []AuthorityAdoptionObject,
) error {
	if len(policy.Roles) != 1 || !policy.TerminateSessions ||
		!policy.RequireZeroRemainingSessions || !policy.RestoreBeforeCommit ||
		policy.ExpectedPreNamespaceSHA256 != AuthorityOperationPreNamespaceSHA256(objects) ||
		policy.ExpectedPostNamespaceSHA256 != AuthorityOperationPostNamespaceSHA256(objects) {
		return contractError("authority quiescence policy")
	}
	role := policy.Roles[0]
	if role.Principal != authorityQuiescedPrincipal ||
		!validAuthorityMigratorAtRest(role.PreState) ||
		role.TemporaryState != authorityMigratorTemporaryState(role.PreState) ||
		role.PostState != role.PreState ||
		!role.PasswordVerifierMustRemainUnchanged {
		return contractError("authority quiescence role policy")
	}
	return nil
}

func validateAuthorityQuiescenceEvidence(
	evidence AuthorityQuiescenceEvidence, objects []AuthorityAdoptionReceiptObject,
) error {
	if len(evidence.Roles) != 1 || !evidence.RestoredBeforeCommit ||
		!validDigest(evidence.ObservedPreNamespaceSHA256) ||
		evidence.ObservedPostNamespaceSHA256 != AuthorityReceiptNamespaceSHA256(objects) {
		return contractError("authority quiescence evidence")
	}
	role := evidence.Roles[0]
	if role.Principal != authorityQuiescedPrincipal ||
		!validAuthorityMigratorAtRest(role.ObservedPreState) ||
		role.ObservedTemporaryState != authorityMigratorTemporaryState(role.ObservedPreState) ||
		role.ObservedPostState != role.ObservedPreState ||
		role.TerminatedSessionCount < 0 || role.TerminatedSessionCount > 1000000 ||
		role.RemainingSessionCount != 0 || !role.PasswordVerifierUnchanged {
		return contractError("authority quiescence role evidence")
	}
	return nil
}

func ValidateAuthorityQuiescenceBinding(
	policy AuthorityQuiescencePolicy, evidence AuthorityQuiescenceEvidence,
) error {
	if len(policy.Roles) != 1 || len(evidence.Roles) != 1 ||
		evidence.ObservedPreNamespaceSHA256 != policy.ExpectedPreNamespaceSHA256 ||
		evidence.ObservedPostNamespaceSHA256 != policy.ExpectedPostNamespaceSHA256 ||
		evidence.RestoredBeforeCommit != policy.RestoreBeforeCommit {
		return contractError("authority quiescence binding")
	}
	planned, observed := policy.Roles[0], evidence.Roles[0]
	if observed.Principal != planned.Principal ||
		observed.ObservedPreState != planned.PreState ||
		observed.ObservedTemporaryState != planned.TemporaryState ||
		observed.ObservedPostState != planned.PostState ||
		observed.PasswordVerifierUnchanged != planned.PasswordVerifierMustRemainUnchanged {
		return contractError("authority quiescence role binding")
	}
	return nil
}

func authorityMigratorAtRestState(passwordState string) AuthorityQuiescedRoleState {
	return AuthorityQuiescedRoleState{
		Login: true, ConnectionLimit: -1, PasswordState: passwordState,
		ValidUntil: "infinity", MembershipCount: 1,
	}
}

func authorityMigratorTemporaryState(atRest AuthorityQuiescedRoleState) AuthorityQuiescedRoleState {
	state := atRest
	state.Login = false
	state.ConnectionLimit = 0
	return state
}

func validAuthorityMigratorAtRest(state AuthorityQuiescedRoleState) bool {
	return state.Login && !state.Inherit && !state.Superuser && !state.CreateDB &&
		!state.CreateRole && !state.Replication && !state.BypassRLS &&
		state.ConnectionLimit == -1 && state.ValidUntil == "infinity" &&
		state.MembershipCount == 1 &&
		(state.PasswordState == "null" || state.PasswordState == "scram_sha_256")
}
