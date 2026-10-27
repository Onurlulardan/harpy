package iam

type permissionRequestCreateDocument struct {
	Mode      string                                `json:"mode"`
	ExpiresAt string                                `json:"expires_at"`
	Items     []permissionRequestCreateItemDocument `json:"items"`
}

type permissionRequestCreateItemDocument struct {
	TargetKind   string `json:"target_kind"`
	TargetID     string `json:"target_id"`
	PermissionID string `json:"permission_id"`
}

type permissionRequestStateDocument struct {
	ExpectedState string `json:"expected_state"`
}

type permissionRequestDecisionDocument struct {
	ExpectedState string `json:"expected_state"`
	ReviewTaskID  string `json:"review_task_id"`
}
