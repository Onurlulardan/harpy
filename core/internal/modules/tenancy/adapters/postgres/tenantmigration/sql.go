package tenantmigration

import (
	"embed"
	"strings"
)

//go:embed sql/platform/*.sql
var files embed.FS

func load(name string) string {
	content, err := files.ReadFile("sql/platform/" + name + ".sql")
	if err != nil {
		panic("tenant migration runtime SQL is unavailable")
	}
	return strings.TrimSpace(string(content))
}

var (
	currentReadySQL = load("current_ready")
	enqueueSQL      = load("enqueue")
	getSQL          = load("get")
)
