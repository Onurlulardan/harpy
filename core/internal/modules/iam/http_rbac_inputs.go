package iam

type rbacCatalogDocument struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
}

type rbacPermissionCreateDocument struct {
	ResourceID     string   `json:"resource_id"`
	Target         string   `json:"target"`
	UserID         *string  `json:"user_id"`
	RoleID         *string  `json:"role_id"`
	OrganizationID *string  `json:"organization_id"`
	ActionIDs      []string `json:"action_ids"`
}

type rbacPermissionUpdateDocument struct {
	ResourceID     *string   `json:"resource_id"`
	Target         *string   `json:"target"`
	UserID         *string   `json:"user_id"`
	RoleID         *string   `json:"role_id"`
	OrganizationID *string   `json:"organization_id"`
	ActionIDs      *[]string `json:"action_ids"`
}

type rbacCheckDocument struct {
	ResourceSlug string `json:"resource_slug"`
	ActionSlug   string `json:"action_slug"`
}

type rbacBatchCheckDocument struct {
	Checks []rbacCheckDocument `json:"checks"`
}
