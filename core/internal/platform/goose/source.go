package goose

import "embed"

//go:embed sql/platform
var platformRoot embed.FS

//go:embed sql/tenant
var tenantRoot embed.FS
