package postgres

import _ "embed"

var (
	//go:embed sql/platform/record.sql
	recordSQL string
	//go:embed sql/platform/overview.sql
	overviewSQL string
	//go:embed sql/platform/active_users.sql
	activeUsersSQL string
	//go:embed sql/platform/top_users.sql
	topUsersSQL string
	//go:embed sql/platform/user_summary.sql
	userSummarySQL string
	//go:embed sql/platform/user_timeline.sql
	userTimelineSQL string
	//go:embed sql/platform/features.sql
	featuresSQL string
	//go:embed sql/platform/heatmap.sql
	heatmapSQL string
	//go:embed sql/platform/trend.sql
	trendSQL string
	//go:embed sql/platform/retention_prune.sql
	retentionPruneSQL string
)
