package storagecatalog

type Manifest struct {
	Version int     `json:"version"`
	Owners  []Owner `json:"owners"`
}

type Owner struct {
	Name              string                  `json:"name"`
	Prefix            string                  `json:"prefix"`
	Scopes            []string                `json:"scopes"`
	MigrationDirs     []DirectoryClaim        `json:"migration_dirs"`
	StaticSQLDirs     []DirectoryClaim        `json:"static_sql_dirs"`
	PlatformSchemas   []string                `json:"platform_schemas"`
	SchemaUsageGrants []SchemaGrant           `json:"schema_usage_grants"`
	PlatformRoutines  []string                `json:"platform_routines"`
	RoutineExecGrants []RoutineGrant          `json:"routine_execute_grants"`
	PlatformTables    []string                `json:"platform_tables"`
	ExternalAccesses  []ExternalAccess        `json:"external_accesses"`
	ForeignKeys       []ForeignKey            `json:"foreign_keys"`
	LegacyTables      []LegacyTable           `json:"legacy_tables"`
	DynamicNamespaces []DynamicNamespaceClaim `json:"dynamic_namespaces"`
}

type SchemaGrant struct {
	Schema   string `json:"schema"`
	Role     string `json:"role"`
	Contract string `json:"contract"`
}

type RoutineGrant struct {
	Routine           string `json:"routine"`
	IdentityArguments string `json:"identity_arguments"`
	Role              string `json:"role"`
	Contract          string `json:"contract"`
}

type ExternalAccess struct {
	Access      string `json:"access"`
	TargetOwner string `json:"target_owner"`
	Target      string `json:"target"`
	Contract    string `json:"contract"`
}

type ForeignKey struct {
	Constraint        string   `json:"constraint"`
	Source            string   `json:"source"`
	SourceColumns     []string `json:"source_columns"`
	TargetOwner       string   `json:"target_owner"`
	Target            string   `json:"target"`
	TargetColumns     []string `json:"target_columns"`
	MatchType         string   `json:"match_type"`
	OnUpdate          string   `json:"on_update"`
	OnDelete          string   `json:"on_delete"`
	Deferrable        bool     `json:"deferrable"`
	InitiallyDeferred bool     `json:"initially_deferred"`
	Contract          string   `json:"contract"`
}

type DirectoryClaim struct {
	Path  string `json:"path"`
	Scope string `json:"scope"`
}

type DynamicNamespaceClaim struct {
	Kind      string `json:"kind"`
	Prefix    string `json:"prefix"`
	Authority string `json:"authority"`
}

type LegacyTable struct {
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	Adapter     string `json:"adapter"`
	CutoverGate string `json:"cutover_gate"`
}
