package iam

type RBACResourceView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	OwnerDomain string `json:"owner_domain"`
	Kind        string `json:"resource_kind"`
	Protected   bool   `json:"is_system"`
	State       string `json:"state"`
	Revision    int64  `json:"revision"`
}

type RBACActionView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Protected   bool   `json:"is_system"`
	State       string `json:"state"`
	Revision    int64  `json:"revision"`
}

type RBACPermissionView struct {
	ID             string           `json:"id"`
	Target         string           `json:"target"`
	ResourceID     string           `json:"resource_id"`
	UserID         *string          `json:"user_id,omitempty"`
	RoleID         *string          `json:"role_id,omitempty"`
	OrganizationID *string          `json:"organization_id,omitempty"`
	Resource       RBACResourceView `json:"resource"`
	Actions        []RBACActionView `json:"actions"`
	State          string           `json:"state"`
	Revision       int64            `json:"revision"`
}

type RBACResourcePage struct {
	Items   []RBACResourceView
	HasMore bool
}

type RBACActionPage struct {
	Items   []RBACActionView
	HasMore bool
}

type RBACPermissionPage struct {
	Items       []RBACPermissionView
	HasMore     bool
	NextAfterID *string
}

type RBACDecisionView struct {
	Allowed              bool   `json:"allowed"`
	Reason               string `json:"reason"`
	AuthorizationClock   int64  `json:"authorization_clock"`
	AuthorizationVersion int64  `json:"authorization_version"`
}

type RBACBatchDecisionView struct {
	Results map[string]bool `json:"results"`
}

type RBACEffectiveView struct {
	PermissionID string  `json:"permission_id"`
	MatchKind    string  `json:"match_kind"`
	ResourceID   *string `json:"resource_id,omitempty"`
	ResourceSlug string  `json:"resource_slug"`
	ActionID     *string `json:"action_id,omitempty"`
	ActionSlug   string  `json:"action_slug"`
	Source       string  `json:"source"`
	SourceID     string  `json:"source_id"`
}

type RBACEffectivePage struct {
	Items []RBACEffectiveView `json:"items"`
}
