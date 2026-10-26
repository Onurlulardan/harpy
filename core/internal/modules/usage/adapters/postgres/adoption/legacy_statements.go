package adoption

import _ "embed"

var (
	//go:embed sql/legacy/manifest.sql
	legacyManifestSQL string
	//go:embed sql/legacy/page.sql
	legacyPageSQL string
)
