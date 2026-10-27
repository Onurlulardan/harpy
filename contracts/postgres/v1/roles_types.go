package postgresv1

type RolesManifest struct {
	ContractVersion string         `json:"contract_version"`
	DatabaseOrigin  string         `json:"database_origin"`
	Roles           []PostgresRole `json:"roles"`
}

type PostgresRole struct {
	Name            string `json:"name"`
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
}
