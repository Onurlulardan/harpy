package cutoverv1

type AuthorityQuiescencePolicy struct {
	Roles                        []AuthorityQuiescedRolePolicy `json:"roles"`
	TerminateSessions            bool                          `json:"terminate_sessions"`
	RequireZeroRemainingSessions bool                          `json:"require_zero_remaining_sessions"`
	RestoreBeforeCommit          bool                          `json:"restore_before_commit"`
	ExpectedPreNamespaceSHA256   string                        `json:"expected_pre_namespace_sha256"`
	ExpectedPostNamespaceSHA256  string                        `json:"expected_post_namespace_sha256"`
}

type AuthorityQuiescedRolePolicy struct {
	Principal                           string                     `json:"principal"`
	PreState                            AuthorityQuiescedRoleState `json:"pre_state"`
	TemporaryState                      AuthorityQuiescedRoleState `json:"temporary_state"`
	PostState                           AuthorityQuiescedRoleState `json:"post_state"`
	PasswordVerifierMustRemainUnchanged bool                       `json:"password_verifier_must_remain_unchanged"`
}

type AuthorityQuiescedRoleState struct {
	Login           bool   `json:"login"`
	Inherit         bool   `json:"inherit"`
	Superuser       bool   `json:"superuser"`
	CreateDB        bool   `json:"createdb"`
	CreateRole      bool   `json:"createrole"`
	Replication     bool   `json:"replication"`
	BypassRLS       bool   `json:"bypassrls"`
	ConnectionLimit int    `json:"connection_limit"`
	PasswordState   string `json:"password_state"`
	ValidUntil      string `json:"valid_until"`
	MembershipCount int    `json:"membership_count"`
}

type AuthorityQuiescenceEvidence struct {
	Roles                       []AuthorityQuiescedRoleEvidence `json:"roles"`
	ObservedPreNamespaceSHA256  string                          `json:"observed_pre_namespace_sha256"`
	ObservedPostNamespaceSHA256 string                          `json:"observed_post_namespace_sha256"`
	RestoredBeforeCommit        bool                            `json:"restored_before_commit"`
}

type AuthorityQuiescedRoleEvidence struct {
	Principal                 string                     `json:"principal"`
	ObservedPreState          AuthorityQuiescedRoleState `json:"observed_pre_state"`
	ObservedTemporaryState    AuthorityQuiescedRoleState `json:"observed_temporary_state"`
	ObservedPostState         AuthorityQuiescedRoleState `json:"observed_post_state"`
	TerminatedSessionCount    int                        `json:"terminated_session_count"`
	RemainingSessionCount     int                        `json:"remaining_session_count"`
	PasswordVerifierUnchanged bool                       `json:"password_verifier_unchanged"`
}
