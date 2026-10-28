package migrationbundle

type corpusSet struct {
	owner, scope, root string
	globs              []string
	supersede          bool
}

func corpusSets() []corpusSet {
	return []corpusSet{
		{"platform_migrator", "platform", "core/internal/platform/migrator", []string{"migrations/fresh/*.sql"}, false},
		{"platform_cutover", "platform", "core/internal/platform/cutover", []string{"migrations/fresh/*.sql", "migrations/supersessions/fresh/*.sql"}, true},
		{"tenancy", "platform", "core/internal/modules/tenancy", []string{"migrations/platform/fresh/*.sql", "migrations/platform/supersessions/fresh/*.sql"}, false},
		{"tenancy", "tenant", "core/internal/modules/tenancy", []string{"migrations/tenant/fresh/*.sql"}, false},
		{"platformauthority", "platform", "core/internal/modules/platformauthority", []string{"migrations/platform/fresh/*.fresh.sql"}, false},
		{"platformops", "platform", "core/internal/modules/platformops", []string{"migrations/platform/fresh/*.fresh.sql"}, false},
		{"iam", "platform", "core/internal/modules/iam", []string{"migrations/platform/fresh/*.sql"}, false},
		{"iam", "tenant", "core/internal/modules/iam", []string{"migrations/tenant/fresh/*.sql", "migrations/tenant/supersessions/fresh/*.sql"}, false},
		{"system_resource_layout", "tenant", "core/internal/modules/system_resource_layout", []string{"migrations/tenant/fresh/*.fresh.sql", "migrations/tenant/supersessions/fresh/*.fresh.sql", "migrations/tenant/supersessions/v2/fresh/*.fresh.sql", "migrations/tenant/supersessions/v3/fresh/*.fresh.sql"}, false},
		{"billing", "platform", "core/internal/modules/billing", []string{"migrations/platform/fresh/*.sql"}, false},
		{"usage", "platform", "core/internal/modules/usage", []string{"migrations/platform/fresh/*.fresh.sql"}, false},
		{"platform_async", "platform", "core/internal/platform/async", []string{"migrations/fresh/*.sql", "migrations/supersessions/fresh/*.sql"}, true},
		{"platform_backupreceipt", "platform", "core/internal/platform/backupreceipt", []string{"migrations/fresh/*.sql"}, false},
		{"analytics", "platform", "core/internal/modules/analytics", []string{"migrations/platform/fresh/*.sql"}, false},
		{"analytics", "tenant", "core/internal/modules/analytics", []string{"migrations/tenant/fresh/*.sql", "migrations/tenant/supersessions/fresh/*.sql"}, false},
		{"catalog", "platform", "core/internal/modules/catalog", []string{"migrations/platform/fresh/*.sql"}, false},
		{"catalog", "tenant", "core/internal/modules/catalog", []string{"migrations/tenant/fresh/*.sql", "migrations/tenant/supersessions/fresh/*.sql"}, false},
		{"work", "tenant", "core/internal/modules/work", []string{"migrations/tenant/fresh/*.fresh.sql"}, false},
		{"documents", "tenant", "core/internal/modules/documents", []string{"migrations/tenant/fresh/*.fresh.sql"}, false},
		{"notifications", "platform", "core/internal/modules/notifications", []string{"migrations/platform/fresh/*.fresh.sql"}, false},
		{"notifications", "tenant", "core/internal/modules/notifications", []string{"migrations/tenant/fresh/*.fresh.sql"}, false},
		{"support", "platform", "core/internal/modules/support", []string{"migrations/platform/fresh/*.fresh.sql"}, false},
		{"audit", "tenant", "core/internal/modules/audit", []string{"migrations/tenant/fresh/*.fresh.sql"}, false},
	}
}
