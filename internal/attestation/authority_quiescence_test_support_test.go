package attestation

import cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"

func authorityQuiescencePolicyFixture(
	objects []cutoverv1.AuthorityAdoptionObject,
) cutoverv1.AuthorityQuiescencePolicy {
	atRest := authorityMigratorAtRestFixture()
	temporary := atRest
	temporary.Login = false
	temporary.ConnectionLimit = 0
	return cutoverv1.AuthorityQuiescencePolicy{
		Roles: []cutoverv1.AuthorityQuiescedRolePolicy{{
			Principal: "harpy_migrator", PreState: atRest,
			TemporaryState: temporary, PostState: atRest,
			PasswordVerifierMustRemainUnchanged: true,
		}},
		TerminateSessions: true, RequireZeroRemainingSessions: true,
		RestoreBeforeCommit:         true,
		ExpectedPreNamespaceSHA256:  cutoverv1.AuthorityOperationPreNamespaceSHA256(objects),
		ExpectedPostNamespaceSHA256: cutoverv1.AuthorityOperationPostNamespaceSHA256(objects),
	}
}

func authorityQuiescenceEvidenceFixture(
	operation cutoverv1.AuthorityAdoptionOperation,
	objects []cutoverv1.AuthorityAdoptionReceiptObject,
) cutoverv1.AuthorityQuiescenceEvidence {
	planned := operation.QuiescencePolicy.Roles[0]
	return cutoverv1.AuthorityQuiescenceEvidence{
		Roles: []cutoverv1.AuthorityQuiescedRoleEvidence{{
			Principal: planned.Principal, ObservedPreState: planned.PreState,
			ObservedTemporaryState:    planned.TemporaryState,
			ObservedPostState:         planned.PostState,
			PasswordVerifierUnchanged: true,
		}},
		ObservedPreNamespaceSHA256:  operation.QuiescencePolicy.ExpectedPreNamespaceSHA256,
		ObservedPostNamespaceSHA256: cutoverv1.AuthorityReceiptNamespaceSHA256(objects),
		RestoredBeforeCommit:        true,
	}
}

func authorityMigratorAtRestFixture() cutoverv1.AuthorityQuiescedRoleState {
	return cutoverv1.AuthorityQuiescedRoleState{
		Login: true, ConnectionLimit: -1, PasswordState: "scram_sha_256",
		ValidUntil: "infinity", MembershipCount: 1,
	}
}
