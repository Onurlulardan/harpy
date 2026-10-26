package postgresv1

type MigrationManifest struct {
	ContractVersion string           `json:"contract_version"`
	Engine          string           `json:"engine"`
	PostgresMajor   int              `json:"postgres_major"`
	HistoryTable    string           `json:"history_table"`
	OutOfOrder      bool             `json:"out_of_order"`
	Entries         []MigrationEntry `json:"entries"`
}

type MigrationEntry struct {
	Owner                        string            `json:"owner"`
	Scope                        string            `json:"scope"`
	Version                      int64             `json:"version"`
	TransactionMode              string            `json:"transaction_mode"`
	Phase                        string            `json:"phase"`
	TargetDatabaseGeneration     int64             `json:"target_database_generation"`
	NextMinimumRuntimeGeneration int64             `json:"next_minimum_runtime_generation"`
	Variants                     MigrationVariants `json:"variants"`
}

type MigrationVariants struct {
	Fresh     *MigrationVariant `json:"fresh,omitempty" harpyOptional:"allowed"`
	V1Adopted *MigrationVariant `json:"v1_adopted,omitempty" harpyOptional:"allowed"`
}

type MigrationVariant struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type SelectedMigration struct {
	Owner                        string
	Scope                        string
	Version                      int64
	TransactionMode              string
	Phase                        string
	TargetDatabaseGeneration     int64
	NextMinimumRuntimeGeneration int64
	Origin                       DatabaseOrigin
	Path                         string
	SHA256                       string
}
