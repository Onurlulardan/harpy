package postgres

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed sql/platform/*.sql
var statements embed.FS

func statement(name string) string {
	content, err := statements.ReadFile("sql/platform/" + name + ".sql")
	if err != nil {
		panic(fmt.Sprintf("tenant migration SQL %s is unavailable", name))
	}
	return strings.TrimSpace(string(content))
}

var (
	registerReadySQL  = statement("register_ready")
	heartbeatReadySQL = statement("heartbeat_ready")
	retireReadySQL    = statement("retire_ready")
	claimSQL          = statement("claim")
	heartbeatSQL      = statement("heartbeat")
	completeSQL       = statement("complete")
	failSQL           = statement("fail")
	inspectSchemaSQL  = statement("inspect_schema")
	recordSchemaSQL   = statement("record_schema")
)
