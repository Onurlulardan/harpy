package storagecatalog

const (
	rootDocumentPath      = "contracts/storage/owners.yaml"
	ownerDirectoryPath    = "contracts/storage/owners"
	fragmentDirectoryPath = "contracts/storage/fragments"
	catalogVersion        = 2
	inlineOwnerVersion    = 1
	modularOwnerVersion   = 2
	ownerFragmentVersion  = 1
	maximumIncludes       = 256
	maximumRootBytes      = 64 * 1024
	maximumFragmentBytes  = 1024 * 1024
	maximumCatalogBytes   = 16 * 1024 * 1024
	ownerDocumentVersion  = inlineOwnerVersion
)

type rootDocument struct {
	Version  int      `json:"version"`
	Includes []string `json:"includes"`
}

type inlineOwnerDocument struct {
	Version int   `json:"version"`
	Owner   Owner `json:"owner"`
}

type ownerDocument = inlineOwnerDocument

type modularOwnerDocument struct {
	Version  int      `json:"version"`
	Owner    Owner    `json:"owner"`
	Includes []string `json:"includes"`
}

type documentVersion struct {
	Version int `json:"version"`
}

type ownerFragmentDocument struct {
	Version           int                     `json:"version"`
	Owner             string                  `json:"owner"`
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
