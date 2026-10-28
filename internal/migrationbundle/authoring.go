package migrationbundle

// Fresh authoring stamps every current SQL file with the same phase and
// generation. Those values match the fresh binary schema contract. They are
// not a per-file rollout classification, and they do not rewrite archived
// release manifests.
const (
	FreshAuthoringPhase              = "expand"
	FreshAuthoringDatabaseGeneration = int64(1)
	FreshAuthoringRuntimeGeneration  = int64(1)
	freshManifestContract            = "postgres-migration-manifest.v1"
	freshManifestEngine              = "goose.v3"
	freshManifestHistory             = "platform.platform_migration_history"
	freshManifestPostgresMajor       = 18
)

// Source is one fresh SQL file. History-supersession chains are not part of
// this value; runtime documents keep those chains separately.
type Source struct {
	Owner string
	Scope string
	Path  string
	Body  []byte
}
