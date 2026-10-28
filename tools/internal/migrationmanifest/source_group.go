package migrationmanifest

type SourceGroup struct {
	Owner                        string
	Scope                        string
	Phase                        string
	TargetDatabaseGeneration     int64
	NextMinimumRuntimeGeneration int64
	VersionFrom                  int64
	VersionTo                    int64
	Directory                    string
	Documents                    []SourceDocument
}
