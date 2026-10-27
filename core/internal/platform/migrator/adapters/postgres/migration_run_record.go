package postgres

type MigrationRunRecord struct {
	RunID                    string
	ReleaseID                string
	EnvironmentID            string
	DatabaseOrigin           string
	Command                  string
	State                    string
	MigrationManifestSHA256  string
	ExtensionManifestSHA256  string
	RolesManifestSHA256      string
	PrivilegesManifestSHA256 string
	OwnershipManifestSHA256  string
	FailureCode              string
}

type CreateMigrationRunInput struct {
	RunID                    string
	ReleaseID                string
	EnvironmentID            string
	DatabaseOrigin           string
	Command                  string
	MigrationManifestSHA256  string
	ExtensionManifestSHA256  string
	RolesManifestSHA256      string
	PrivilegesManifestSHA256 string
	OwnershipManifestSHA256  string
}
