package postgresv1

type PrivilegesManifest struct {
	ContractVersion          string              `json:"contract_version"`
	DatabaseOrigin           string              `json:"database_origin"`
	DatabaseName             string              `json:"database_name"`
	PublicDatabasePrivileges []string            `json:"public_database_privileges"`
	PublicSchemaCreate       bool                `json:"public_schema_create"`
	Memberships              []RoleMembership    `json:"memberships"`
	Grants                   []ACLGrant          `json:"grants"`
	DefaultPrivileges        []DefaultPrivilege  `json:"default_privileges"`
	ExecutorDecisions        ExecutorDecisionSet `json:"executor_decisions"`
}

type RoleMembership struct {
	Member        string `json:"member"`
	Role          string `json:"role"`
	AdminOption   bool   `json:"admin_option"`
	InheritOption bool   `json:"inherit_option"`
	SetOption     bool   `json:"set_option"`
}

type CatalogObject struct {
	Kind              string    `json:"kind"`
	DatabaseName      string    `json:"database_name"`
	SchemaName        *string   `json:"schema_name,omitempty" harpyOptional:"allowed"`
	ObjectName        string    `json:"object_name"`
	IdentityArguments *[]string `json:"identity_arguments,omitempty" harpyOptional:"allowed"`
}

type ACLGrant struct {
	Object      CatalogObject `json:"object"`
	Grantor     string        `json:"grantor"`
	Grantee     string        `json:"grantee"`
	Privileges  []string      `json:"privileges"`
	GrantOption bool          `json:"grant_option"`
}

type DefaultPrivilege struct {
	Owner       string   `json:"owner"`
	SchemaName  string   `json:"schema_name"`
	ObjectKind  string   `json:"object_kind"`
	Grantee     string   `json:"grantee"`
	Privileges  []string `json:"privileges"`
	GrantOption bool     `json:"grant_option"`
}

type ExecutorDecisionSet struct {
	BootstrapRole       string `json:"bootstrap_role"`
	ExtensionRole       string `json:"extension_role"`
	RuntimeSessionRole  string `json:"runtime_session_role"`
	RuntimeCurrentRole  string `json:"runtime_current_role"`
	MigratorSessionRole string `json:"migrator_session_role"`
	MigratorCurrentRole string `json:"migrator_current_role"`
	OwnershipExecutor   string `json:"ownership_executor"`
}
