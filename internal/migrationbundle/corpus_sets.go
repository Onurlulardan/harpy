package migrationbundle

type corpusSet struct {
	owner, scope, root string
	globs              []string
	supersede          bool
}

func corpusSets() []corpusSet {
	return []corpusSet{
		{"platform_baseline", "platform", "core/internal/platform/baseline", []string{"sql/20261028000000_platform_baseline.fresh.sql"}, false},
		{"tenant_baseline", "tenant", "core/internal/platform/baseline", []string{"sql/20261028000001_tenant_baseline.fresh.sql"}, false},
	}
}
